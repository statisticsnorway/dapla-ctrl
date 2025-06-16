import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { UserProfile } from '../@types/user'
import { Effect, Option as O } from 'effect'

import { customLogger } from '../utils/logger.ts'

type UserProfileStoreState = {
  loggedInUser: O.Option<UserProfile>
}

type UserProfileStoreActions = {
  setLoggedInUser: (user: UserProfile) => void
}

export type UserProfileStore = UserProfileStoreState & UserProfileStoreActions

export const useUserProfileStore = create<UserProfileStore>()(
  subscribeWithSelector((set) => ({
    loggedInUser: O.none(),
    setLoggedInUser: (user: UserProfile) => set(() => ({ loggedInUser: O.some(user) })),
  }))
)

useUserProfileStore.subscribe(
  (state) => state.loggedInUser,
  (user) =>
    Effect.log('Retrieving logged in user from store:', O.getOrNull(user)).pipe(
      Effect.provide(customLogger),
      Effect.runPromise
    )
)
