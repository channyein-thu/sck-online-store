
import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { ApprovePointDto, CreatePointDto, RedeemPointDto } from '../point.dto';
import { Point, PointStatus } from '../point.entity';
import { InsufficientPointsError, PointNotFoundError, PointService } from '../point.service';

describe('PointService', () => {
  let service: PointService;

  const mockPointRepository = {
    save: jest.fn(),
    find: jest.fn(),
    findOne: jest.fn(),
    update: jest.fn(),
  };

  beforeEach(() => {
    mockPointRepository.update.mockResolvedValue({ affected: 0 });
  });

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        PointService,
        {
          provide: getRepositoryToken(Point),
          useValue: mockPointRepository,
        },
      ],
    }).compile();

    service = module.get<PointService>(PointService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('Should be defined', () => {
    expect(service).toBeDefined();
  });

  it('Create => Should create a new point with amount computed from amountThb and return its data', async () => {
    // arrange
    const createPointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      amountThb: 10000,
    } as CreatePointDto;

    const expectedSaveArg = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      amount: 200,
    };

    const createPointResponse = {
      id: 1,
      orgId: 1,
      userId: 1,
      orderId: 10,
      amount: 200,
      status: PointStatus.PENDING_APPROVAL,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    } as unknown as Point;

    jest.spyOn(mockPointRepository, 'save').mockReturnValue(createPointResponse);

    // act
    const result = await service.deductPoint(createPointInput);

    // assert
    expect(mockPointRepository.save).toBeCalled();
    expect(mockPointRepository.save).toBeCalledWith(expectedSaveArg);
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

    jest.spyOn(mockPointRepository, 'find').mockReturnValue(points);

    //act
    const result = await service.getPoint();

    // assert
    expect(result).toEqual(points);
    expect(mockPointRepository.find).toBeCalled();
  });

  it('Find => should flip overdue APPROVED points to EXPIRED before returning', async () => {
    // arrange
    jest.spyOn(mockPointRepository, 'find').mockReturnValue([]);

    // act
    await service.getPoint();

    // assert
    expect(mockPointRepository.update).toBeCalledWith(
      { status: PointStatus.APPROVED, expiryDate: expect.anything() },
      { status: PointStatus.EXPIRED },
    );
  });

  describe('getBalance', () => {
    it('Should sum the amount of APPROVED points that are not expired', async () => {
      // arrange
      const points = [
        { id: 1, orgId: 1, userId: 1, amount: 100, status: PointStatus.APPROVED },
        { id: 2, orgId: 1, userId: 1, amount: 50, status: PointStatus.APPROVED },
      ] as unknown as Point[];

      jest.spyOn(mockPointRepository, 'find').mockResolvedValue(points);

      // act
      const result = await service.getBalance(1, 1);

      // assert
      expect(mockPointRepository.find).toBeCalledWith({
        where: [
          {
            orgId: 1,
            userId: 1,
            status: PointStatus.APPROVED,
            expiryDate: expect.anything(),
          },
          {
            orgId: 1,
            userId: 1,
            status: PointStatus.APPROVED,
            expiryDate: expect.anything(),
          },
          {
            orgId: 1,
            userId: 1,
            status: PointStatus.REDEEMED,
          },
        ],
      });
      expect(result).toEqual({ point: 150 });
    });

    it('Should net out REDEEMED points against APPROVED points regardless of expiry', async () => {
      // arrange
      const points = [
        { id: 1, orgId: 1, userId: 1, amount: 9, status: PointStatus.APPROVED },
        { id: 2, orgId: 1, userId: 1, amount: 25, status: PointStatus.APPROVED },
        { id: 3, orgId: 1, userId: 1, amount: -9, status: PointStatus.REDEEMED },
      ] as unknown as Point[];

      jest.spyOn(mockPointRepository, 'find').mockResolvedValue(points);

      // act
      const result = await service.getBalance(1, 1);

      // assert
      expect(result).toEqual({ point: 25 });
    });

    it('Should return 0 when there are no matching points', async () => {
      // arrange
      jest.spyOn(mockPointRepository, 'find').mockResolvedValue([]);

      // act
      const result = await service.getBalance(1, 1);

      // assert
      expect(result).toEqual({ point: 0 });
    });
  });

  describe('approvePoint', () => {
    const approvePointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
    } as ApprovePointDto;

    it('Should approve a PENDING_APPROVAL point and set approvedAt/expiryDate to +179 days', async () => {
      // arrange
      const pendingPoint = {
        id: 1,
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: 200,
        status: PointStatus.PENDING_APPROVAL,
        approvedAt: null,
        expiryDate: null,
        created: '2024-08-25T09:06:58',
        updated: '2024-08-25T09:06:58',
      } as unknown as Point;

      jest.spyOn(mockPointRepository, 'findOne').mockResolvedValueOnce(pendingPoint);
      jest.spyOn(mockPointRepository, 'save').mockImplementation((entity) => entity);

      // act
      const result = await service.approvePoint(approvePointInput);

      // assert
      expect(mockPointRepository.findOne).toBeCalledWith({
        where: {
          orgId: 1,
          userId: 1,
          orderId: 10,
          status: PointStatus.PENDING_APPROVAL,
        },
      });
      expect(result.status).toEqual(PointStatus.APPROVED);
      expect(result.approvedAt).toBeInstanceOf(Date);
      expect(result.expiryDate).toBeInstanceOf(Date);

      const diffInDays = Math.round(
        (result.expiryDate.getTime() - result.approvedAt.getTime()) / (24 * 60 * 60 * 1000),
      );
      expect(diffInDays).toEqual(179);
    });

    it('Should return the already-approved record as-is (idempotent) when no PENDING_APPROVAL record exists', async () => {
      // arrange
      const approvedPoint = {
        id: 1,
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: 200,
        status: PointStatus.APPROVED,
        approvedAt: new Date('2026-01-01T00:00:00Z'),
        expiryDate: new Date('2026-06-29T00:00:00Z'),
        created: '2024-08-25T09:06:58',
        updated: '2024-08-25T09:06:58',
      } as unknown as Point;

      jest.spyOn(mockPointRepository, 'findOne').mockResolvedValueOnce(null);
      jest.spyOn(mockPointRepository, 'findOne').mockResolvedValueOnce(approvedPoint);

      // act
      const result = await service.approvePoint(approvePointInput);

      // assert
      expect(mockPointRepository.save).not.toBeCalled();
      expect(result).toEqual(approvedPoint);
    });

    it('Should throw PointNotFoundError when no point record exists for the order at all', async () => {
      // arrange
      jest.spyOn(mockPointRepository, 'findOne').mockResolvedValueOnce(null);
      jest.spyOn(mockPointRepository, 'findOne').mockResolvedValueOnce(null);

      // act & assert
      await expect(service.approvePoint(approvePointInput)).rejects.toThrow(PointNotFoundError);
      expect(mockPointRepository.save).not.toBeCalled();
    });
  });

  describe('redeemPoint', () => {
    const redeemPointInput = {
      orgId: 1,
      userId: 1,
      orderId: 10,
      points: 50,
    } as RedeemPointDto;

    it('Should redeem points and insert a negative-amount REDEEMED row when balance is sufficient', async () => {
      // arrange
      const approvedPoints = [
        { id: 1, orgId: 1, userId: 1, amount: 100, status: PointStatus.APPROVED },
      ] as unknown as Point[];

      jest.spyOn(mockPointRepository, 'find').mockResolvedValue(approvedPoints);
      jest.spyOn(mockPointRepository, 'save').mockImplementation((entity) => ({
        id: 2,
        ...entity,
      }));

      // act
      const result = await service.redeemPoint(redeemPointInput);

      // assert
      expect(mockPointRepository.save).toBeCalledWith({
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: -50,
        status: PointStatus.REDEEMED,
      });
      expect(result).toEqual({
        id: 2,
        orgId: 1,
        userId: 1,
        orderId: 10,
        amount: -50,
        status: PointStatus.REDEEMED,
      });
    });

    it('Should throw InsufficientPointsError without inserting when requested points exceed the balance', async () => {
      // arrange
      const approvedPoints = [
        { id: 1, orgId: 1, userId: 1, amount: 20, status: PointStatus.APPROVED },
      ] as unknown as Point[];

      jest.spyOn(mockPointRepository, 'find').mockResolvedValue(approvedPoints);

      // act & assert
      await expect(service.redeemPoint(redeemPointInput)).rejects.toThrow(InsufficientPointsError);
      expect(mockPointRepository.save).not.toBeCalled();
    });
  });

});