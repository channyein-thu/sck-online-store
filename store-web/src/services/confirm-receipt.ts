import axiosShoppingMallApi from '@/utils/axios'
import { handleServiceError } from '@/utils/helper'

// ------------------------------------------------

export type ConfirmReceiptServiceResponse = {
  data?: {
    order_number: number
    status: string
  }
  message?: string
}

const confirmReceiptService = async (
  orderNumber: number
): Promise<ConfirmReceiptServiceResponse> => {
  try {
    const { data } = await axiosShoppingMallApi.post(
      `/api/v1/order/${orderNumber}/confirmReceipt`
    )
    return {
      data: data
    }
  } catch (error) {
    return handleServiceError(error)
  }
}

export default confirmReceiptService
