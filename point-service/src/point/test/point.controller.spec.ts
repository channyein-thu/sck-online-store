
import { HttpStatus } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import { ApprovePointDto, CreatePointDto, RedeemPointDto } from '../point.dto';
import { PointController } from '../point.controller';
import { InsufficientPointsError, PointNotFoundError, PointService } from '../point.service';

describe('PointController', () => {
  let controller: PointController;

  const mockPointService = {
    getPoint: jest.fn(),
    deductPoint: jest.fn(),
    approvePoint: jest.fn(),
    getBalance: jest.fn(),
    redeemPoint: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [PointController],
      providers: [
        {
          provide: PointService,
          useValue: mockPointService,
        },
      ],
    }).compile();

    controller = module.get<PointController>(PointController);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should be defined', () => {
    expect(controller).toBeDefined();
  });

  it('Create => should create a new point by a given data', async () => {
    // arrange
    const createPointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      amountThb: 10000,
    } as CreatePointDto;

    const createPointResponse = {
      id: 1,
      orgId: 1,
      userId: 1,
      orderId: 10,
      amount: 200,
      status: 'PENDING_APPROVAL',
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    };


    jest.spyOn(mockPointService, 'deductPoint').mockReturnValue(createPointResponse);

    // act
    const result = await controller.createPoint(createPointInput);

    // assert
    expect(mockPointService.deductPoint).toBeCalled();
    expect(mockPointService.deductPoint).toBeCalledWith(createPointInput);
    expect(result).toEqual(createPointResponse);
  });

  it('Find => should return an array of point', async () => {
    //arrange
    const point = {
      id: 2,
      orgId: 1,
      userId: 1,
      amount: 300,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    };
    const points = [point];
    jest.spyOn(mockPointService, 'getPoint').mockReturnValue(points);

    //act
    const result = await controller.getPoint();

    // assert
    expect(result).toEqual(points);
    expect(mockPointService.getPoint).toBeCalled();
  });

  describe('getBalance', () => {
    it('Should call PointService.getBalance with numeric orgId/userId and return its result', async () => {
      // arrange
      jest.spyOn(mockPointService, 'getBalance').mockResolvedValue({ point: 150 });

      // act
      const result = await controller.getBalance('1', '1');

      // assert
      expect(mockPointService.getBalance).toBeCalledWith(1, 1);
      expect(result).toEqual({ point: 150 });
    });

    it('Should map an unexpected error to a 500 HttpException', async () => {
      // arrange
      jest.spyOn(mockPointService, 'getBalance').mockRejectedValue(new Error('DB is down'));

      // act & assert
      await expect(controller.getBalance('1', '1')).rejects.toMatchObject({
        status: HttpStatus.INTERNAL_SERVER_ERROR,
      });
    });
  });

  describe('approvePoint', () => {
    const approvePointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
    } as ApprovePointDto;

    it('Should call PointService.approvePoint and return its result', async () => {
      // arrange
      const approvePointResponse = {
        id: 1,
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: 200,
        status: 'APPROVED',
        approvedAt: '2026-01-01T00:00:00.000Z',
        expiryDate: '2026-06-29',
        created: '2024-08-25T09:06:58',
        updated: '2024-08-25T09:06:58',
      };

      jest.spyOn(mockPointService, 'approvePoint').mockResolvedValue(approvePointResponse);

      // act
      const result = await controller.approvePoint(approvePointInput);

      // assert
      expect(mockPointService.approvePoint).toBeCalledWith(approvePointInput);
      expect(result).toEqual(approvePointResponse);
    });

    it('Should map PointNotFoundError to a 404 HttpException with a NOT_FOUND code', async () => {
      // arrange
      const notFoundError = new PointNotFoundError('No point record found for orgId=1, userId=1, orderId=10');
      jest.spyOn(mockPointService, 'approvePoint').mockRejectedValue(notFoundError);

      // act & assert
      await expect(controller.approvePoint(approvePointInput)).rejects.toMatchObject({
        status: HttpStatus.NOT_FOUND,
        response: {
          code: 'NOT_FOUND',
          message: notFoundError.message,
        },
      });
    });

    it('Should map an unexpected error to a 500 HttpException', async () => {
      // arrange
      jest.spyOn(mockPointService, 'approvePoint').mockRejectedValue(new Error('DB is down'));

      // act & assert
      await expect(controller.approvePoint(approvePointInput)).rejects.toMatchObject({
        status: HttpStatus.INTERNAL_SERVER_ERROR,
      });
    });
  });

  describe('redeemPoint', () => {
    const redeemPointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      points: 50,
    } as RedeemPointDto;

    it('Should call PointService.redeemPoint and return its result', async () => {
      // arrange
      const redeemPointResponse = {
        id: 2,
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: -50,
        status: 'REDEEMED',
      };

      jest.spyOn(mockPointService, 'redeemPoint').mockResolvedValue(redeemPointResponse);

      // act
      const result = await controller.redeemPoint(redeemPointInput);

      // assert
      expect(mockPointService.redeemPoint).toBeCalledWith(redeemPointInput);
      expect(result).toEqual(redeemPointResponse);
    });

    it('Should map InsufficientPointsError to a 400 HttpException with an INSUFFICIENT_POINTS code', async () => {
      // arrange
      const insufficientError = new InsufficientPointsError(
        'Insufficient points: orgId=1, userId=1, requested=50, balance=20',
      );
      jest.spyOn(mockPointService, 'redeemPoint').mockRejectedValue(insufficientError);

      // act & assert
      await expect(controller.redeemPoint(redeemPointInput)).rejects.toMatchObject({
        status: HttpStatus.BAD_REQUEST,
        response: {
          code: 'INSUFFICIENT_POINTS',
          message: insufficientError.message,
        },
      });
    });

    it('Should map an unexpected error to a 500 HttpException', async () => {
      // arrange
      jest.spyOn(mockPointService, 'redeemPoint').mockRejectedValue(new Error('DB is down'));

      // act & assert
      await expect(controller.redeemPoint(redeemPointInput)).rejects.toMatchObject({
        status: HttpStatus.INTERNAL_SERVER_ERROR,
      });
    });
  });

});