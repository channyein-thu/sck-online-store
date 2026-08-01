'use client'

import Button from '@/components/button/button'
import Header2 from '@/components/typography/header2'
import Text from '@/components/typography/text'
import confirmReceiptService from '@/services/confirm-receipt'
import getOrderHistoryService, {
  OrderHistoryItem
} from '@/services/order-history'
import { converNumber } from '@/utils/format'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'

// ----------------------------------------------------------------------

const HistoryView = () => {
  const [orderList, setOrderList] = useState<OrderHistoryItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [confirmingOrderNumbers, setConfirmingOrderNumbers] = useState<
    number[]
  >([])
  const [confirmErrors, setConfirmErrors] = useState<Record<number, string>>(
    {}
  )

  useEffect(() => {
    const getOrderHistory = async () => {
      const result = await getOrderHistoryService()

      if (result.data) {
        setOrderList(result.data)
      } else {
        setOrderList([])
      }
      setIsLoading(false)
    }

    getOrderHistory()
  }, [])

  const handleConfirmReceipt = async (orderNumber: number) => {
    setConfirmingOrderNumbers((prev) => [...prev, orderNumber])
    setConfirmErrors((prev) => {
      const next = { ...prev }
      delete next[orderNumber]
      return next
    })

    const result = await confirmReceiptService(orderNumber)

    if (result.data) {
      setOrderList((prev) =>
        prev.map((order) =>
          order.order_number === orderNumber
            ? { ...order, status: 'completed' }
            : order
        )
      )
    } else {
      setConfirmErrors((prev) => ({
        ...prev,
        [orderNumber]:
          result.message || 'Something went wrong, please try again'
      }))
    }

    setConfirmingOrderNumbers((prev) =>
      prev.filter((number) => number !== orderNumber)
    )
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-10">
      <Header2 id="order-history-title">Order History</Header2>

      {isLoading ? (
        <Text id="order-history-loading">Loading your orders...</Text>
      ) : orderList.length === 0 ? (
        <Text id="order-history-empty">You have no orders yet</Text>
      ) : (
        <div className="overflow-x-auto border border-gray-200 rounded-md">
          <table id="order-history-table" className="min-w-full divide-y divide-gray-200">
            <thead>
              <tr className="text-left text-sm font-medium text-gray-900">
                <th className="px-4 py-3">Order Number</th>
                <th className="px-4 py-3">Date</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Total Price</th>
                <th className="px-4 py-3">Earn Point</th>
                <th className="px-4 py-3">Burn Point</th>
                <th className="px-4 py-3">Tracking Number</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {orderList.map((order) => {
                const isConfirming = confirmingOrderNumbers.includes(
                  order.order_number
                )
                const confirmError = confirmErrors[order.order_number]

                return (
                  <tr
                    id={`order-history-row-${order.order_number}`}
                    key={order.order_number}
                    className="text-sm text-gray-600"
                  >
                    <td className="px-4 py-3">{order.order_number}</td>
                    <td className="px-4 py-3">
                      {dayjs(order.updated).format('DD/MM/YYYY HH:mm:ss')}
                    </td>
                    <td className="px-4 py-3">{order.status}</td>
                    <td className="px-4 py-3">
                      {converNumber(order.total_price)}
                    </td>
                    <td className="px-4 py-3">
                      {converNumber(order.earn_point)}
                    </td>
                    <td className="px-4 py-3">
                      {converNumber(order.burn_point)}
                    </td>
                    <td className="px-4 py-3">{order.tracking_no || '-'}</td>
                    <td className="px-4 py-3">
                      {order.status === 'paid' && (
                        <>
                          <Button
                            id={`order-history-confirm-${order.order_number}`}
                            type="button"
                            size="sm"
                            disabled={isConfirming}
                            className="disabled:opacity-50 disabled:cursor-not-allowed"
                            onClick={() =>
                              handleConfirmReceipt(order.order_number)
                            }
                          >
                            {isConfirming
                              ? 'Confirming...'
                              : 'Confirm Receipt'}
                          </Button>
                          {confirmError ? (
                            <Text
                              id={`order-history-confirm-error-${order.order_number}`}
                              size="sm"
                              className="text-red-600 mt-1"
                            >
                              {confirmError}
                            </Text>
                          ) : null}
                        </>
                      )}
                      {order.status === 'completed' && (
                        <Text
                          id={`order-history-confirmed-${order.order_number}`}
                          size="sm"
                        >
                          Confirmed
                        </Text>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

export default HistoryView
