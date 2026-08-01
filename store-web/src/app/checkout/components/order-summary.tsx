'use client'

import SummaryText from '@/app/checkout/components/summary-text'
import Header3 from '@/components/typography/header3'
import Text from '@/components/typography/text'
import useOrderStore from '@/hooks/use-order-store'
import { convertCurrency } from '@/utils/format'

// ----------------------------------------------------------------------

const OrderSummary = () => {
  const { summary, totalPayment, receivePoint, shipping, point } =
    useOrderStore((state) => state)

  const pointDiscount = Math.floor(point.burnPoint / 2)
  const hasDiscount = point.burnPoint > 0
  const originalTotal = summary.total_price_thb + shipping.shippingFee

  return (
    <div className="mb-6">
      <Header3>Summary</Header3>

      <div className="mb-6 pb-2 border-b border-gray-200 text-gray-800">
        <SummaryText
          id="order-summary-subtotal"
          text="Merchandise Subtotal"
          value={summary.total_price_thb}
        />
        <SummaryText
          id="order-summary-receive-point"
          text="Receive Points"
          format="number"
          className="font-semibold"
          unit="Points"
          value={receivePoint}
        />
        <SummaryText
          id="order-summary-shipping-fee"
          text="Shipping Fee"
          value={shipping.shippingFee}
        />
        {hasDiscount && (
          <SummaryText
            id="order-summary-point-discount"
            text="Points Discount"
            textBeforeValue="-"
            format="currency"
            className="text-red-600 font-semibold"
            value={pointDiscount}
          />
        )}

        {/* Not use for now */}
        {/* <SummaryText text='Discount' value={20.0} />
        <SummaryText text='Tax (7%)' value={3.99} /> */}
      </div>
      <div>
        {hasDiscount && (
          <div className="w-full flex justify-end mb-1">
            <Text
              id="order-summary-original-total"
              size="sm"
              className="line-through text-gray-400"
            >
              {convertCurrency(originalTotal, 'THB')}
            </Text>
          </div>
        )}
        <SummaryText
          id="order-summary-total-payment"
          className="font-semibold"
          text="Total Payment"
          value={totalPayment}
        />
      </div>
    </div>
  )
}

export default OrderSummary
