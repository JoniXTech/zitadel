import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { NewFeatureService } from '../services/new-feature.service';

/**
 * Route guard that checks whether user self-service (profile page) is enabled.
 * When the instance feature `disableUserSelfService` is enabled, normal users
 * are redirected away from their profile page (/users/me).
 */
export const userSelfServiceGuard: CanActivateFn = async () => {
  const featureService = inject(NewFeatureService);
  const router = inject(Router);

  try {
    const features = await featureService.getInstanceFeatures();
    if (features.disableUserSelfService?.enabled) {
      router.navigate(['/']);
      return false;
    }
  } catch {
    // If features can't be fetched, allow access by default
  }
  return true;
};
