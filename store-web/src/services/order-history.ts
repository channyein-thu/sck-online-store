import axiosShoppingMallApi from '@/utils/axios'
import { handleServiceError } from '@/utils/helper'

// ------------------------------------------------

export type OrderHistoryItem = {
  order_number: number
  status: string
  sub_total_price: number
  total_price: number
  burn_point: number
  earn_point: number
  tracking_no: string
  updated: string
}

export type GetOrderHistoryServiceResponse = {
  data?: OrderHistoryItem[]
  message?: string
}

const getOrderHistoryService = async (): Promise<GetOrderHistoryServiceResponse> => {
  try {
    const { data } = await axiosShoppingMallApi.get(`/api/v1/order/history`)
    return {
      data: data
    }
  } catch (error) {
    return handleServiceError(error)
  }
}

export default getOrderHistoryService
