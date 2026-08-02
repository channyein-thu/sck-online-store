
import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { ApprovePointDto, CreatePointDto } from '../point.dto';
import { Point, PointStatus } from '../point.entity';
import { PointNotFoundError, PointService } from '../point.service';

describe('PointService', () => {
  let service: PointService;

  const mockPointRepository = {
    save: jest.fn(),
    find: jest.fn(),
    findOne: jest.fn(),
  };

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

  it('Create => Should create a new point and return its data', async () => {
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

    jest.spyOn(mockPointRepository, 'save').mockReturnValue(createPointResponse);

    // act
    const result = await service.deductPoint(createPointInput);

    // assert
    expect(mockPointRepository.save).toBeCalled();
    expect(mockPointRepository.save).toBeCalledWith(createPointInput);
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

  describe('calculatePoint', () => {
    it('Should floor priceThb/50 into whole points', () => {
      expect(service.calculatePoint(647.5)).toEqual({ point: 12 });
    });

    it('Should return 0 points when priceThb is below the rate', () => {
      expect(service.calculatePoint(49)).toEqual({ point: 0 });
    });
  });

  describe('getBalance', () => {
    it('Should split points into pendingApproval, approved, redeemed, and expired buckets', async () => {
      // arrange
      const today = new Date();
      const future = new Date(today);
      future.setDate(future.getDate() + 30);
      const past = new Date(today);
      past.setDate(past.getDate() - 30);

      const points = [
        { id: 1, orgId: 1, userId: 1, amount: 200, status: PointStatus.PENDING_APPROVAL, expiryDate: null },
        { id: 2, orgId: 1, userId: 1, amount: 100, status: PointStatus.APPROVED, expiryDate: future },
        { id: 3, orgId: 1, userId: 1, amount: 50, status: PointStatus.APPROVED, expiryDate: null },
        { id: 4, orgId: 1, userId: 1, amount: -30, status: PointStatus.REDEEMED, expiryDate: null },
        { id: 5, orgId: 1, userId: 1, amount: 40, status: PointStatus.APPROVED, expiryDate: past },
      ] as unknown as Point[];

      jest.spyOn(mockPointRepository, 'find').mockResolvedValue(points);

      // act
      const result = await service.getBalance(1, 1);

      // assert
      expect(mockPointRepository.find).toBeCalledWith({ where: { orgId: 1, userId: 1 } });
      expect(result).toEqual([
        { status: PointStatus.PENDING_APPROVAL, point: 200 },
        { status: PointStatus.APPROVED, point: 150 },
        { status: PointStatus.REDEEMED, point: 30 },
        { status: PointStatus.EXPIRED, point: 40 },
      ]);
    });

    it('Should return all zeros when there are no points for the user', async () => {
      // arrange
      jest.spyOn(mockPointRepository, 'find').mockResolvedValue([]);

      // act
      const result = await service.getBalance(1, 1);

      // assert
      expect(result).toEqual([
        { status: PointStatus.PENDING_APPROVAL, point: 0 },
        { status: PointStatus.APPROVED, point: 0 },
        { status: PointStatus.REDEEMED, point: 0 },
        { status: PointStatus.EXPIRED, point: 0 },
      ]);
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

});
