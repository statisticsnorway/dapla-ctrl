import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { User } from '../services/userProfile.ts'
import { Effect, Option as O } from 'effect'

import { customLogger } from '../utils/logger.ts'

type UserProfileStoreState = {
  loggedInUser: O.Option<User>
}

type UserProfileStoreActions = {
  setUser: (user: User) => void
}

export type UserProfileStore = UserProfileStoreState & UserProfileStoreActions

export const useUserProfileStore = create<UserProfileStore>()(
  subscribeWithSelector((set) => ({
    loggedInUser: O.none(),
    setUser: (user: User) => set(() => ({ loggedInUser: O.some(user) })),
  }))
)

useUserProfileStore.subscribe(
  (state) => state.loggedInUser,
  (user) => Effect.log('USER LOGGED IN:', O.getOrNull(user)).pipe(Effect.provide(customLogger), Effect.runPromise)
)
