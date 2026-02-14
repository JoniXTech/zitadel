import { DynamicTheme } from "@/components/dynamic-theme";
import { SignInWithIdp } from "@/components/sign-in-with-idp";
import { Translated } from "@/components/translated";
import { UsernameForm } from "@/components/username-form";
import { idpTypeToSlug } from "@/lib/idp";
import { getServiceConfig } from "@/lib/service-url";
import { getPublicHost } from "@/lib/server/host";
import {
  getActiveIdentityProviders,
  getBrandingSettings,
  getDefaultOrg,
  getLoginSettings,
  startIdentityProviderFlow,
} from "@/lib/zitadel";
import { IdentityProviderType } from "@zitadel/proto/zitadel/settings/v2/login_settings_pb";
import { Organization } from "@zitadel/proto/zitadel/org/v2/org_pb";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("loginname");
  return { title: t("title") };
}

export default async function Page(props: { searchParams: Promise<Record<string | number | symbol, string | undefined>> }) {
  const searchParams = await props.searchParams;

  const loginName = searchParams?.loginName;
  const requestId = searchParams?.requestId;
  const organization = searchParams?.organization;
  const suffix = searchParams?.suffix;
  const submit: boolean = searchParams?.submit === "true";
  const idpHint = searchParams?.idp_hint;

  const _headers = await headers();
  const { serviceConfig } = getServiceConfig(_headers);

  let defaultOrganization;
  if (!organization) {
    const org: Organization | null = await getDefaultOrg({ serviceConfig });
    if (org) {
      defaultOrganization = org.id;
    }
  }

  const loginSettings = await getLoginSettings({ serviceConfig, organization: organization ?? defaultOrganization });

  const identityProviders = await getActiveIdentityProviders({
    serviceConfig,
    orgId: organization ?? defaultOrganization,
  }).then((resp) => {
    return resp.identityProviders;
  });

  const branding = await getBrandingSettings({ serviceConfig, organization: organization ?? defaultOrganization });

  // If idp_hint is provided, auto-redirect to the matching identity provider
  if (idpHint && loginSettings?.allowExternalIdp && identityProviders?.length) {
    const matchedIdp =
      identityProviders.find((idp) => idp.id === idpHint) ||
      identityProviders.find((idp) => idp.name.toLowerCase() === idpHint.toLowerCase());

    if (matchedIdp) {
      const host = getPublicHost(_headers);
      const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? "";
      const protocol = host.includes("localhost") ? "http://" : "https://";

      if (matchedIdp.type === IdentityProviderType.LDAP) {
        const ldapParams = new URLSearchParams();
        if (requestId) ldapParams.set("requestId", requestId);
        if (organization) ldapParams.set("organization", organization);
        ldapParams.set("idpId", matchedIdp.id);
        redirect(`/idp/ldap?${ldapParams.toString()}`);
      }

      const provider = idpTypeToSlug(matchedIdp.type);
      const params = new URLSearchParams();
      if (requestId) params.set("requestId", requestId);
      if (organization) params.set("organization", organization);

      const url = await startIdentityProviderFlow({
        serviceConfig,
        idpId: matchedIdp.id,
        urls: {
          successUrl: `${protocol}${host}${basePath}/idp/${provider}/process?${params.toString()}`,
          failureUrl: `${protocol}${host}${basePath}/idp/${provider}/failure?${params.toString()}`,
        },
      });

      if (url) {
        redirect(url);
      }
    }
  }

  return (
    <DynamicTheme branding={branding}>
      <div className="flex flex-col space-y-4">
        <h1>
          <Translated i18nKey="title" namespace="loginname" />
        </h1>
        <p className="ztdl-p">
          <Translated i18nKey="description" namespace="loginname" />
        </p>
      </div>

      <div className="w-full">
        <UsernameForm
          loginName={loginName}
          requestId={requestId}
          organization={organization} // stick to "organization" as we still want to do user discovery based on the searchParams not the default organization, later the organization is determined by the found user
          defaultOrganization={defaultOrganization}
          loginSettings={loginSettings}
          suffix={suffix}
          submit={submit}
          allowRegister={!!loginSettings?.allowRegister}
        ></UsernameForm>

        {loginSettings?.allowExternalIdp && !!identityProviders?.length && (
          <div className="w-full pb-4 pt-6">
            <SignInWithIdp
              identityProviders={identityProviders}
              requestId={requestId}
              organization={organization}
              postErrorRedirectUrl="/loginname"
            ></SignInWithIdp>
          </div>
        )}
      </div>
    </DynamicTheme>
  );
}
