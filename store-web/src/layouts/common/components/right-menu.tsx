'use client'

import { useUserStore } from '@/hooks/use-user-store'
import Cart from '@/layouts/common/components/cart'
import Login from '@/layouts/common/components/login'
import PointBadge from '@/layouts/common/components/point-badge'
import { HeaderProps } from '@/layouts/common/header'
import { UserCircleIcon } from '@heroicons/react/16/solid'
// ---------------------------------------------------

const RightMenu = ({ setShoppingCartOpen }: HeaderProps) => {
  const user = useUserStore((state) => state.user)
  return (
    <div className="flex flex-1 gap-x-10 justify-end">
      <Cart setShoppingCartOpen={setShoppingCartOpen} />
      {user ? (
        <div className="flex justify-center items-center gap-1">
          <PointBadge />
          <UserCircleIcon className="h-7 w-7 text-gray-800" />
          {/* <span>{user.firstName.toLocaleUpperCase()}</span> */}
        </div>
      ) : (
        <Login />
      )}
    </div>
  )
}

export default RightMenu
