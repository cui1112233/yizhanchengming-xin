const scalarText = value => value === undefined || value === null ? '' : String(value).trim();

export const websiteProfileOrganization = profile => {
  const advanced = profile?.advanced || profile?.config_data || {};
  return scalarText(
    profile?.organization
    ?? profile?.organization_id
    ?? profile?.organizationId
    ?? profile?.org_id
    ?? advanced?.organization_id
    ?? advanced?.organizationId
    ?? advanced?.organization
    ?? advanced?.org_id
  );
};

export function reconcileSyncedWebsiteProfiles(publishSettings, syncedProfiles) {
  const current = publishSettings && typeof publishSettings === 'object' ? publishSettings : {};
  const profiles = Array.isArray(syncedProfiles) ? syncedProfiles : [];
  const currentId = scalarText(current.websiteProfileId);
  const selected = profiles.find(profile => scalarText(profile?.id) === currentId);
  const selectionExpired = Boolean(currentId && !selected);

  return {
    publishSettings: {
      ...current,
      websiteProfiles: profiles,
      websiteProfileId: selected ? currentId : '',
      versionProfile: selected ? scalarText(selected.name) : '',
      organization: selected ? websiteProfileOrganization(selected) : ''
    },
    selectionExpired
  };
}
