import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { IsNull, LessThan, MoreThanOrEqual, Repository } from 'typeorm';
import { Point, PointStatus } from './point.entity';
import { ApprovePointDto, CreatePointDto, RedeemPointDto } from './point.dto';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';

const otelLogger = logs.getLogger('point-service');

const VALIDITY_DAYS = 179; // confirmation date counts as Day 1 of the 180-day window

export class PointNotFoundError extends Error {}
export class InsufficientPointsError extends Error {}

@Injectable()
export class PointService {
  private readonly logger = new Logger(PointService.name);

  constructor(
    @InjectRepository(Point)
    private pointRepository: Repository<Point>,
  ) {}

  private async expireOverduePoints(): Promise<void> {
    const today = new Date();
    const result = await this.pointRepository.update(
      { status: PointStatus.APPROVED, expiryDate: LessThan(today) },
      { status: PointStatus.EXPIRED },
    );

    if (result.affected) {
      this.logger.log(`Points expired: count=${result.affected}`);
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points expired',
        attributes: {
          'log_type': 'state_change',
          'event': 'points_expired',
          'entity_type': 'point',
          'count': result.affected,
        },
      });
    }
  }

  async getPoint(): Promise<Point[]> {
    try {
      await this.expireOverduePoints();
      const points = await this.pointRepository.find();
      this.logger.log(`Points retrieved, count=${points.length}`);
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points retrieved',
        attributes: {
          'log_type': 'business',
          'event': 'points_retrieved',
          'entity_type': 'point',
          'items_count': points.length,
        },
      });
      return points;
    } catch (error) {
      this.logger.error('PointRepository.find internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.find internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }

  private async computeBalance(orgId: number, userId: number): Promise<number> {
    await this.expireOverduePoints();

    const today = new Date();
    const points = await this.pointRepository.find({
      where: [
        {
          orgId,
          userId,
          status: PointStatus.APPROVED,
          expiryDate: IsNull(),
        },
        {
          orgId,
          userId,
          status: PointStatus.APPROVED,
          expiryDate: MoreThanOrEqual(today),
        },
        {
          orgId,
          userId,
          status: PointStatus.REDEEMED,
        },
      ],
    });

    return points.reduce((sum, item) => sum + item.amount, 0);
  }

  async getBalance(orgId: number, userId: number): Promise<{ point: number }> {
    try {
      const balance = await this.computeBalance(orgId, userId);

      this.logger.log(
        `Balance retrieved: orgId=${orgId}, userId=${userId}, point=${balance}`,
      );
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Balance retrieved',
        attributes: {
          'log_type': 'business',
          'event': 'balance_retrieved',
          'entity_type': 'point',
          'org_id': orgId,
          'user_id': userId,
          'point': balance,
        },
      });

      return { point: balance };
    } catch (error) {
      this.logger.error('PointRepository.getBalance internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.getBalance internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }

  async deductPoint(input: CreatePointDto): Promise<Point> {
    try {
      // 50 THB = 1 point, hardcoded for now — configurable rate is a later step
      const amount = Math.floor(input.amountThb / 50);
      const saved = await this.pointRepository.save({
        orgId: input.orgId,
        userId: input.userId,
        orderId: input.orderId,
        amount,
      });
      this.logger.log(
        `Points deducted: userId=${input.userId}, orgId=${input.orgId}, orderId=${input.orderId}, amount=${amount}`,
      );
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points deducted',
        attributes: {
          'log_type': 'state_change',
          'event': 'points_deducted',
          'entity_type': 'point',
          'entity_id': saved.id,
          'changed_by': input.userId,
          'org_id': input.orgId,
          'order_id': input.orderId,
          'amount': amount,
          'status': saved.status,
        },
      });
      return saved;
    } catch (error) {
      this.logger.error('PointRepository.save internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.save internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }

  async approvePoint(input: ApprovePointDto): Promise<Point> {
    try {
      const pending = await this.pointRepository.findOne({
        where: {
          orgId: input.orgId,
          userId: input.userId,
          orderId: input.orderId,
          status: PointStatus.PENDING_APPROVAL,
        },
      });

      if (pending) {
        const approvedAt = new Date();
        const expiryDate = new Date(approvedAt);
        expiryDate.setDate(expiryDate.getDate() + VALIDITY_DAYS);

        const approved = await this.pointRepository.save({
          ...pending,
          status: PointStatus.APPROVED,
          approvedAt,
          expiryDate,
        });

        this.logger.log(
          `Point approved: userId=${input.userId}, orgId=${input.orgId}, orderId=${input.orderId}, expiryDate=${expiryDate.toISOString()}`,
        );
        otelLogger.emit({
          severityNumber: SeverityNumber.INFO,
          severityText: 'INFO',
          body: 'Point approved',
          attributes: {
            'log_type': 'state_change',
            'event': 'point_approved',
            'entity_type': 'point',
            'entity_id': approved.id,
            'changed_by': input.userId,
            'org_id': input.orgId,
            'order_id': input.orderId,
            'approved_at': approvedAt.toISOString(),
            'expiry_date': expiryDate.toISOString(),
          },
        });
        return approved;
      }

      const alreadyApproved = await this.pointRepository.findOne({
        where: {
          orgId: input.orgId,
          userId: input.userId,
          orderId: input.orderId,
          status: PointStatus.APPROVED,
        },
      });

      if (alreadyApproved) {
        this.logger.log(
          `Point already approved, returning as-is: userId=${input.userId}, orgId=${input.orgId}, orderId=${input.orderId}`,
        );
        return alreadyApproved;
      }

      throw new PointNotFoundError(
        `No point record found for orgId=${input.orgId}, userId=${input.userId}, orderId=${input.orderId}`,
      );
    } catch (error) {
      if (error instanceof PointNotFoundError) {
        throw error;
      }
      this.logger.error('PointRepository.approvePoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.approvePoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }

  async redeemPoint(input: RedeemPointDto): Promise<Point> {
    try {
      const balance = await this.computeBalance(input.orgId, input.userId);

      if (input.points > balance) {
        throw new InsufficientPointsError(
          `Insufficient points: orgId=${input.orgId}, userId=${input.userId}, requested=${input.points}, balance=${balance}`,
        );
      }

      const saved = await this.pointRepository.save({
        orgId: input.orgId,
        userId: input.userId,
        orderId: input.orderId,
        amount: -input.points,
        status: PointStatus.REDEEMED,
      });

      this.logger.log(
        `Points redeemed: userId=${input.userId}, orgId=${input.orgId}, orderId=${input.orderId}, points=${input.points}`,
      );
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points redeemed',
        attributes: {
          'log_type': 'state_change',
          'event': 'points_redeemed',
          'entity_type': 'point',
          'entity_id': saved.id,
          'changed_by': input.userId,
          'org_id': input.orgId,
          'order_id': input.orderId,
          'amount': -input.points,
          'status': saved.status,
        },
      });

      return saved;
    } catch (error) {
      if (error instanceof InsufficientPointsError) {
        throw error;
      }
      this.logger.error('PointRepository.redeemPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.redeemPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }
}
