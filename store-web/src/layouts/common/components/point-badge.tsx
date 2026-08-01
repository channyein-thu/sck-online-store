'use client'

import getPointService from '@/services/point'
import { converNumber } from '@/utils/format'
import { SparklesIcon } from '@heroicons/react/24/solid'
import { useEffect, useState } from 'react'

// ----------------------------------------------------------------------

const PointBadge = () => {
  const [point, setPoint] = useState<number | null>(null)

  useEffect(() => {
    const getPoint = async () => {
      const result = await getPointService()

      if (result.data) {
        setPoint(result.data.point)
      }
    }

    getPoint()
  }, [])

  if (point === null) {
    return null
  }

  return (
    <div
      id="point-badge"
      className="bg-amber-500 text-white rounded-full px-4 py-2 flex items-center gap-2"
    >
      <SparklesIcon className="h-5 w-5" />
      <span className="font-bold">{`${converNumber(point)} pts`}</span>
      <span>{`= ฿${converNumber(Math.floor(point / 2))}`}</span>
    </div>
  )
}

export default PointBadge
