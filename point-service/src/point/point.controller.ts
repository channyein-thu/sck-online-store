import { Body, Controller, Get, HttpException, HttpStatus, Logger, Post, Query } from '@nestjs/common';
import { InsufficientPointsError, PointNotFoundError, PointService } from './point.service';
import { ApprovePointDto, CreatePointDto, RedeemPointDto } from './point.dto';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';

const otelLogger = logs.getLogger('point-service');

@Controller('point')
export class PointController {
  private readonly logger = new Logger(PointController.name);

  constructor(private readonly pointService: PointService) {}

  @Get()
  async getPoint() {
    this.logger.log('GET /point request received');
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Get points request received',
      attributes: {
        'log_type': 'business',
        'event': 'get_points_request',
        'entity_type': 'point',
      },
    });
    try {
      return await this.pointService.getPoint();
    } catch (error) {
      this.logger.error('PointService.getPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.getPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Get('balance')
  async getBalance(@Query('orgId') orgId: string, @Query('userId') userId: string) {
    this.logger.log(
      `GET /point/balance request received: orgId=${orgId}, userId=${userId}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Get balance request received',
      attributes: {
        'log_type': 'business',
        'event': 'get_balance_request',
        'entity_type': 'point',
        'org_id': orgId,
        'user_id': userId,
      },
    });
    try {
      return await this.pointService.getBalance(Number(orgId), Number(userId));
    } catch (error) {
      this.logger.error('PointService.getBalance internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.getBalance internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Post()
  async createPoint(@Body() body: CreatePointDto) {
    this.logger.log(
      `POST /point request received: userId=${body.userId}, orgId=${body.orgId}, orderId=${body.orderId}, amountThb=${body.amountThb}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Deduct points request received',
      attributes: {
        'log_type': 'business',
        'event': 'deduct_points_request',
        'entity_type': 'point',
        'actor_id': body.userId,
        'org_id': body.orgId,
        'order_id': body.orderId,
        'amount_thb': body.amountThb,
      },
    });
    try {
      return await this.pointService.deductPoint(body);
    } catch (error) {
      this.logger.error('PointService.deductPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.deductPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Post('approve')
  async approvePoint(@Body() body: ApprovePointDto) {
    this.logger.log(
      `POST /point/approve request received: userId=${body.userId}, orgId=${body.orgId}, orderId=${body.orderId}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Approve point request received',
      attributes: {
        'log_type': 'business',
        'event': 'approve_point_request',
        'entity_type': 'point',
        'actor_id': body.userId,
        'org_id': body.orgId,
        'order_id': body.orderId,
      },
    });
    try {
      return await this.pointService.approvePoint(body);
    } catch (error) {
      if (error instanceof PointNotFoundError) {
        this.logger.warn(`PointService.approvePoint not found: ${error.message}`);
        otelLogger.emit({
          severityNumber: SeverityNumber.WARN,
          severityText: 'WARN',
          body: 'PointService.approvePoint not found',
          attributes: { 'error.message': error.message },
        });
        throw new HttpException(
          { code: 'NOT_FOUND', message: error.message },
          HttpStatus.NOT_FOUND,
        );
      }
      this.logger.error('PointService.approvePoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.approvePoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Post('redeem')
  async redeemPoint(@Body() body: RedeemPointDto) {
    this.logger.log(
      `POST /point/redeem request received: userId=${body.userId}, orgId=${body.orgId}, orderId=${body.orderId}, points=${body.points}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Redeem point request received',
      attributes: {
        'log_type': 'business',
        'event': 'redeem_point_request',
        'entity_type': 'point',
        'actor_id': body.userId,
        'org_id': body.orgId,
        'order_id': body.orderId,
        'points': body.points,
      },
    });
    try {
      return await this.pointService.redeemPoint(body);
    } catch (error) {
      if (error instanceof InsufficientPointsError) {
        this.logger.warn(`PointService.redeemPoint insufficient points: ${error.message}`);
        otelLogger.emit({
          severityNumber: SeverityNumber.WARN,
          severityText: 'WARN',
          body: 'PointService.redeemPoint insufficient points',
          attributes: { 'error.message': error.message },
        });
        throw new HttpException(
          { code: 'INSUFFICIENT_POINTS', message: error.message },
          HttpStatus.BAD_REQUEST,
        );
      }
      this.logger.error('PointService.redeemPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.redeemPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }
}
