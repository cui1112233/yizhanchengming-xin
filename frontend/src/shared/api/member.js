import { apiRequest } from './client';

export async function getMemberCenter() {
  const [center, security, notifications] = await Promise.all([
    apiRequest('/api/member/me'),
    apiRequest('/api/member/security').catch(() => null),
    apiRequest('/api/member/notifications?limit=5000').catch(() => null)
  ]);
  return {
    ...center,
    member: center?.member ? { ...center.member, mfaEnabled: Boolean(security?.mfa?.enabled) } : center?.member,
    notificationSummary: notifications ? {
      unread: Number(notifications.unread || 0),
      recent: (notifications.entries || []).slice(0, 8)
    } : center?.notificationSummary
  };
}

export function updateMemberProfile(profile) {
  const payload = typeof profile === 'string' ? { displayName: profile } : (profile || {});
  return apiRequest('/api/member/profile', { method: 'PATCH', body: JSON.stringify(payload) });
}

export function uploadMemberAvatar(avatarDataUrl) {
  return apiRequest('/api/member/avatar', { method: 'POST', body: JSON.stringify({ avatarDataUrl }) });
}

export function uploadMemberWindowBackground(backgroundDataUrl) {
  return apiRequest('/api/member/window-background', { method: 'POST', body: JSON.stringify({ backgroundDataUrl }) });
}

export function getSecurityOverview() { return apiRequest('/api/member/security'); }
export function changeOwnPassword(currentPassword, newPassword) { return apiRequest('/api/member/security/password', { method: 'POST', body: JSON.stringify({ currentPassword, newPassword }) }); }
export function revokeOtherSessions() { return apiRequest('/api/member/security/sessions/revoke-others', { method: 'POST' }); }
export function beginMfaSetup(currentPassword) { return apiRequest('/api/member/security/mfa/setup', { method: 'POST', body: JSON.stringify({ currentPassword }) }); }
export function enableMfa(code) { return apiRequest('/api/member/security/mfa/enable', { method: 'POST', body: JSON.stringify({ code }) }); }
export function disableMfa(currentPassword, code) { return apiRequest('/api/member/security/mfa/disable', { method: 'POST', body: JSON.stringify({ currentPassword, code }) }); }
export function rotateMfaRecoveryCodes(currentPassword, code) { return apiRequest('/api/member/security/mfa/recovery-codes', { method: 'POST', body: JSON.stringify({ currentPassword, code }) }); }

export function getNotifications(limit = 50) { return apiRequest(`/api/member/notifications?limit=${encodeURIComponent(limit)}`); }
export function markNotificationRead(id) { return apiRequest(`/api/member/notifications/${encodeURIComponent(id)}/read`, { method: 'POST' }); }
export function markAllNotificationsRead() { return apiRequest('/api/member/notifications/read-all', { method: 'POST' }); }

export function getTeamMembers() { return apiRequest('/api/member/team'); }
export function getTeamMeta() { return apiRequest('/api/member/team/meta'); }
export function renameTeam(teamId, name) { return apiRequest(`/api/member/team/meta/${encodeURIComponent(teamId)}`, { method: 'PATCH', body: JSON.stringify({ name }) }); }
export function getTeamInvites(manager) { return apiRequest(`/api/member/team/invites${manager ? `?manager=${encodeURIComponent(manager)}` : ''}`); }
export function createTeamInvite(payload) { return apiRequest('/api/member/team/invites', { method: 'POST', body: JSON.stringify(payload) }); }

export function getTeamGovernance() { return apiRequest('/api/member/team/governance'); }
export function updateTeamGovernance(username, monthlyTokenLimit) { return apiRequest(`/api/member/team/governance/${encodeURIComponent(username)}`, { method: 'PATCH', body: JSON.stringify({ monthlyTokenLimit }) }); }

export function createTeamMember(payload) { return apiRequest('/api/member/team/members', { method: 'POST', body: JSON.stringify(payload) }); }
export function updateTeamMember(username, payload) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}`, { method: 'PATCH', body: JSON.stringify(payload) }); }
export function setTeamMemberStatus(username, active) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/status`, { method: 'POST', body: JSON.stringify({ active }) }); }
export function resetTeamMemberPassword(username, password) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/reset-password`, { method: 'POST', body: JSON.stringify({ password }) }); }
export function archiveTeamMember(username, reason = '') { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/archive`, { method: 'POST', body: JSON.stringify({ reason }) }); }
export function restoreArchivedMember(username) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/restore`, { method: 'POST' }); }
export function setTeamMemberApi(username, enabled, scope = '*') { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/api`, { method: 'POST', body: JSON.stringify({ enabled, scope }) }); }
export function setTeamMemberApiScopes(username, scopes) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/api-scopes`, { method: 'PUT', body: JSON.stringify({ scopes }) }); }
export function getTeamMemberUsage(username) { return apiRequest(`/api/member/team/members/${encodeURIComponent(username)}/usage`); }
export function getTeamAudit(limit = 100) { return apiRequest(`/api/member/team/audit?limit=${encodeURIComponent(limit)}`); }

export function inspectTeamInvite(token) { return apiRequest(`/api/login/invite/${encodeURIComponent(token)}`, { suppressGlobalError: true }); }
export function redeemTeamInvite(token, payload) { return apiRequest(`/api/login/invite/${encodeURIComponent(token)}`, { method: 'POST', body: JSON.stringify(payload), suppressGlobalError: true }); }
