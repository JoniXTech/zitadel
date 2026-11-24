import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { GrpcAuthService } from '../services/grpc-auth.service';
import { NewFeatureService } from '../services/new-feature.service';
import { ToastService } from '../services/toast.service';

/**
 * Route guard that checks whether user self-service (profile page) is enabled.
 * When the instance feature `disableUserSelfService` is enabled, normal users
 * are redirected away from their profile page (/users/me).
 * Users with IAM admin permissions (iam.read) bypass this restriction.
 */
export const userSelfServiceGuard: CanActivateFn = async () => {
  const featureService = inject(NewFeatureService);
  const authService = inject(GrpcAuthService);
  const router = inject(Router);
  const toast = inject(ToastService);

  try {
    const features = await featureService.getInstanceFeatures();
    if (features.disableUserSelfService?.enabled) {
      // Allow IAM admins to still access their profile
      const isAdmin = await firstValueFrom(authService.isAllowed(['iam.read']));
      if (isAdmin) {
        return true;
      }
      toast.showError('FEATURES.DISABLEUSERSELFSERVICE_BLOCKED', false, true);
      router.navigate(['/']);
      return false;
    }
  } catch {
    // If features can't be fetched, allow access by default
  }
  return true;
};
