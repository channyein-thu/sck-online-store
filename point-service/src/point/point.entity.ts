// import {
//   Entity,
//   Column,
//   PrimaryGeneratedColumn,
//   CreateDateColumn,
//   UpdateDateColumn,
// } from 'typeorm';

// @Entity('points')
// export class Point {
//   @PrimaryGeneratedColumn()
//   id: number;

//   @Column({ name: 'org_id' })
//   orgId: number;

//   @Column({ name: 'user_id' })
//   userId: number;

//   @Column()
//   amount: number;

//   @CreateDateColumn()
//   created: Date;

//   @UpdateDateColumn()
//   updated: Date;
// }
import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum PointStatus {
  PENDING_APPROVAL = 'PENDING_APPROVAL',
  APPROVED = 'APPROVED',
  REDEEMED = 'REDEEMED',
  EXPIRED = 'EXPIRED',
}

@Entity('points')
export class Point {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ name: 'org_id' })
  orgId: number;

  @Column({ name: 'user_id' })
  userId: number;

  @Column({ name: 'order_id', nullable: true })
  orderId: number;

  @Column()
  amount: number;

  @Column({
    type: 'enum',
    enum: PointStatus,
    default: PointStatus.PENDING_APPROVAL,
  })
  status: PointStatus;

  @Column({ name: 'approved_at', type: 'datetime', nullable: true })
  approvedAt: Date | null;

  @Column({ name: 'expiry_date', type: 'date', nullable: true })
  expiryDate: Date | null;

  @CreateDateColumn()
  created: Date;

  @UpdateDateColumn()
  updated: Date;
}
