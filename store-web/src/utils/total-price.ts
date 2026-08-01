// ----------------------------------------------------------------------------

type SubTotalType = {
  price: number
  quantity: number
}

export const subTotal = (priceList: SubTotalType[]): number => {
  let total = 0

  for (let i = 0; i < priceList.length; i++) {
    total += priceList[i].price * priceList[i].quantity
  }

  return total
}

// 2 points = 1 THB, so the most points that can ever apply to a given
// subtotal is subTotal * 2 (any more would be wasted, not just discounted).
export const pointBurn = (point: number, subTotal: number) => {
  const maxRedeemable = Math.floor(subTotal * 2)

  return Math.min(point, maxRedeemable)
}

export const totalPayment = (
  isUsePoint: boolean,
  pointsUsed: number,
  subTotal: number,
  shippingFee: number
) => {
  let totalPayment = 0

  if (isUsePoint) {
    const discount = Math.floor(pointsUsed / 2)

    if (subTotal <= discount) {
      totalPayment = shippingFee
    } else {
      totalPayment = subTotal - discount + shippingFee
    }
  } else {
    totalPayment = subTotal + shippingFee
  }

  return totalPayment
}
