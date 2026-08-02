import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Point, PointStatus } from './point.entity';
import { ApprovePointDto, CreatePointDto } from './point.dto';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';

const otelLogger = logs.getLogger('point-service');

const VALIDITY_DAYS = 179; // confirmation date counts as Day 1 of the 180-day window

export class PointNotFoundError extends Error {}

@Injectable()
export class PointService {
  private readonly logger = new Logger(PointService.name);

  constructor(
    @InjectRepository(Point)
    private pointRepository: Repository<Point>,
  ) {}

  async getPoint(): Promise<Point[]> {
    try {
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

  calculatePoint(priceThb: number): { point: number } {
    // 50 THB = 1 point, hardcoded for now — configurable rate is a later step
    const point = Math.floor(priceThb / 50);
    this.logger.log(`Point calculated: priceThb=${priceThb}, point=${point}`);
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Point calculated',
      attributes: {
        'log_type': 'business',
        'event': 'point_calculated',
        'entity_type': 'point',
        'price_thb': priceThb,
        'point': point,
      },
    });
    return { point };
  }

  async getBalance(
    orgId: number,
    userId: number,
  ): Promise<{ status: string; point: number }[]> {
    try {
      const points = await this.pointRepository.find({ where: { orgId, userId } });
      // TypeORM hydrates `type: 'date'` columns as 'YYYY-MM-DD' strings despite the
      // `Date` TS type on the entity, so expiryDate must be normalized before comparison
      // (comparing a raw string against a Date coerces the Date to a number and the
      // string to NaN, making every relational comparison silently false).
      const toDateOnly = (d: Date | string): string =>
        d instanceof Date ? d.toISOString().slice(0, 10) : d.slice(0, 10);
      const today = toDateOnly(new Date());

      const pendingApproval = points
        .filter((p) => p.status === PointStatus.PENDING_APPROVAL)
        .reduce((sum, p) => sum + p.amount, 0);

      const approved = points
        .filter(
          (p) =>
            p.status === PointStatus.APPROVED &&
            (!p.expiryDate || toDateOnly(p.expiryDate) >= today),
        )
        .reduce((sum, p) => sum + p.amount, 0);

      const redeemed = Math.abs(
        points
          .filter((p) => p.status === PointStatus.REDEEMED)
          .reduce((sum, p) => sum + p.amount, 0),
      );

      const expired = points
        .filter(
          (p) =>
            p.status === PointStatus.APPROVED && p.expiryDate && toDateOnly(p.expiryDate) < today,
        )
        .reduce((sum, p) => sum + p.amount, 0);

      const balance = [
        { status: PointStatus.PENDING_APPROVAL, point: pendingApproval },
        { status: PointStatus.APPROVED, point: approved },
        { status: PointStatus.REDEEMED, point: redeemed },
        { status: PointStatus.EXPIRED, point: expired },
      ];

      this.logger.log(
        `Balance retrieved: orgId=${orgId}, userId=${userId}, pendingApproval=${pendingApproval}, approved=${approved}, redeemed=${redeemed}, expired=${expired}`,
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
          'pending_approval': pendingApproval,
          'approved': approved,
          'redeemed': redeemed,
          'expired': expired,
        },
      });

      return balance;
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

  async deductPoint(point: CreatePointDto): Promise<Point> {
    try {
      const saved = await this.pointRepository.save(point);
      this.logger.log(
        `Points deducted: userId=${point.userId}, orgId=${point.orgId}, orderId=${point.orderId}, amount=${point.amount}`,
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
          'changed_by': point.userId,
          'org_id': point.orgId,
          'order_id': point.orderId,
          'amount': point.amount,
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
}
