export class CreatePointDto {
  orgId: number;
  userId: number;
  orderId: number;
  amount: number;
}

export class ApprovePointDto {
  orgId: number;
  userId: number;
  orderId: number;
}
