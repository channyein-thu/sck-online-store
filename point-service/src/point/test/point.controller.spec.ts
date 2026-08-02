
import { HttpStatus } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import { ApprovePointDto, CreatePointDto } from '../point.dto';
import { PointController } from '../point.controller';
import { PointNotFoundError, PointService } from '../point.service';

describe('PointController', () => {
  let controller: PointController;

  const mockPointService = {
    getPoint: jest.fn(),
    deductPoint: jest.fn(),
    calculatePoint: jest.fn(),
    approvePoint: jest.fn(),
    getBalance: jest.fn(),
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

  it('should be defined', () => {
    expect(controller).toBeDefined();
  });

  it('Create => should create a new point by a given data', async () => {
    // arrange
    const createPointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      amount: 200,
    } as CreatePointDto;

    const createPointResponse = {
      id: 1,
      orgId: 1,
      userId: 1,
      orderId: 10,
      amount: 200,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    } as CreatePointDto;


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

  it('Calculate => should pass priceThb as a number to PointService.calculatePoint', () => {
    // arrange
    jest.spyOn(mockPointService, 'calculatePoint').mockReturnValue({ point: 12 });

    // act
    const result = controller.calculatePoint('647.5');

    // assert
    expect(mockPointService.calculatePoint).toBeCalledWith(647.5);
    expect(result).toEqual({ point: 12 });
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

});
