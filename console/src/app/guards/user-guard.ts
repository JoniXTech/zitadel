import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map, take, firstValueFrom } from 'rxjs';

import { GrpcAuthService } from '../services/grpc-auth.service';
import { NewFeatureService } from '../services/new-feature.service';

// export const userGuard: CanActivateFn = (route) => {
//   const authService = inject(GrpcAuthService);
//   const router = inject(Router);
//   const featureService = inject(NewFeatureService);

//   return authService.user.pipe(
//     take(1),
//     map(async (user) => {
//       const isMe = user?.id === route.params['id'];
//       if (isMe) {
//         // Don't redirect to /users/me if self-service is disabled
//         try {
//           const features = await featureService.getInstanceFeatures();
//           if (features.disableUserSelfService?.enabled) {
//             return true; // Allow viewing own profile via admin /users/:id route
//           }
//         } catch {
//           // If features can't be fetched, use default behavior
//         }
//         router.navigate(['/users', 'me']).then();
//         return false;
//       }
//       return !isMe;
//     }),
//   );
// };

export const userGuard: CanActivateFn = async (route) => {
  const authService = inject(GrpcAuthService);
  const router = inject(Router);
  const featureService = inject(NewFeatureService);

  const user = await firstValueFrom(authService.user);
  const isMe = user?.id === route.params['id'];
  if (!isMe) {
    return true;
  }
  // Don't redirect to /users/me if self-service is disabled
  try {
    const features = await featureService.getInstanceFeatures();
    if (features.disableUserSelfService?.enabled) {
      return true; // Allow viewing own profile via admin /users/:id route
    }
  } catch {
    // If features can't be fetched, use default redirect behavior
  }
  router.navigate(['/users', 'me']);
  return false;
};
