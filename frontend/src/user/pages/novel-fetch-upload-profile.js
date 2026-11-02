export function mergeDefaultUploadProfile(settings, selectedProfile) {
  return {
    ...(settings && typeof settings === 'object' ? settings : {}),
    selected_profile: String(selectedProfile || '').trim()
  };
}

export function uploadSessionDisplayName(session) {
  if (session?.status !== 'ready') return '未登录';
  return String(session.username || '').trim() || '已登录';
}
