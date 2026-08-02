'use client'

import Text from '@/components/typography/text'
import { convertCurrency, converNumber } from '@/utils/format'

// ----------------------------------------------------------------------

type SubTotalProps = {
  total: number
  point: number
}

const SubTotal = ({ total, point }: SubTotalProps) => {
  return (
    <div className="flex justify-between text-base font-medium text-gray-900">
      <Text id="shopping-cart-subtotal-label">Subtotal</Text>
      <div className="flex flex-col items-end">
        <Text id="shopping-cart-subtotal-price">
          {convertCurrency(total, 'THB')}
        </Text>
        <Text
          id="shopping-cart-subtotal-point"
          size="sm"
          className="text-gray-400"
        >
          {`${converNumber(point)} Points`}
        </Text>
      </div>
    </div>
  )
}

export default SubTotal
