export class CreatePointDto {
  orgId: number;
  userId: number;
  orderId: number;
  amountThb: number;
}

export class ApprovePointDto {
  orgId: number;
  userId: number;
  orderId: number;
}

export class RedeemPointDto {
  orgId: number;
  userId: number;
  orderId: number;
  points: number;
}
