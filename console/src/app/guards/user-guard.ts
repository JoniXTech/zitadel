import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { GrpcAuthService } from '../services/grpc-auth.service';
import { NewFeatureService } from '../services/new-feature.service';

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
      // IAM admins can still access /users/me (the self-service guard handles the bypass)
      const isAdmin = await firstValueFrom(authService.isAllowed(['iam.read']));
      if (isAdmin) {
        router.navigate(['/users', 'me']);
        return false;
      }
      return true; // Non-admin: stay on /users/:id route (self-service guard will block /users/me)
    }
  } catch {
    // If features can't be fetched, use default redirect behavior
  }
  router.navigate(['/users', 'me']);
  return false;
};
