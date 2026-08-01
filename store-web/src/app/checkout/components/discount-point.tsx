'use client'

import Text from '@/components/typography/text'
import useOrderStore from '@/hooks/use-order-store'
import getPointService from '@/services/point'
import { converNumber } from '@/utils/format'
import { pointBurn } from '@/utils/total-price'
import { GiftIcon } from '@heroicons/react/24/outline'
import { useEffect } from 'react'

// ----------------------------------------------------------------------

const DiscountPoint = () => {
  const { point, summary, setPoint, setIsUsePoint, setBurnPoint } =
    useOrderStore((state) => state)

  const maxRedeemable = pointBurn(point.point, summary.total_price_thb)
  const discount = Math.floor(point.burnPoint / 2)

  const handleUsePointChange = (e: { target: { checked: boolean } }) => {
    setIsUsePoint(e.target.checked)
  }

  const handleBurnPointChange = (e: { target: { value: string } }) => {
    setBurnPoint(Number(e.target.value))
  }

  useEffect(() => {
    const getPoint = async () => {
      const result = await getPointService()

      if (result.data) {
        setPoint(result.data.point)
      } else {
        setPoint(0)
      }
    }

    getPoint()
  }, [setPoint])

  return (
    <div
      className={`rounded-lg border p-4 mt-5 transition-colors ${
        point.isUsePoint ? 'border-warning bg-warning/5' : 'border-gray-200'
      }`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <GiftIcon
            className={`h-8 w-8 shrink-0 ${
              point.isUsePoint ? 'text-warning' : 'text-gray-400'
            }`}
          />
          <div>
            <Text
              id="discount-use-point-label"
              size="md"
              className="font-semibold text-gray-900"
            >
              Redeem Points
            </Text>
            <Text id="discount-use-point-total" size="sm">
              {`Available: ${converNumber(point.point)} pts = ฿${converNumber(Math.floor(point.point / 2))}`}
            </Text>
          </div>
        </div>

        <input
          id="discount-use-point-input"
          type="checkbox"
          className="toggle toggle-warning"
          checked={point.isUsePoint}
          onChange={handleUsePointChange}
        />
      </div>

      {point.isUsePoint && (
        <div className="mt-4">
          <input
            id="discount-burn-point-range"
            type="range"
            min={0}
            max={maxRedeemable}
            value={point.burnPoint}
            onChange={handleBurnPointChange}
            className="range range-warning w-full"
          />
          <div className="flex justify-between mt-1">
            <Text id="discount-burn-point-min-label" size="sm">
              0 pts
            </Text>
            <Text id="discount-burn-point-max-label" size="sm">
              {`${converNumber(maxRedeemable)} pts`}
            </Text>
          </div>

          <div className="mt-3 rounded-md border border-warning/30 bg-warning/10 px-3 py-2">
            <Text
              id="discount-point-applied"
              size="sm"
              className="font-semibold text-warning-content"
            >
              {`Discount applied — −฿${converNumber(discount)}`}
            </Text>
          </div>
        </div>
      )}
    </div>
  )
}

export default DiscountPoint
