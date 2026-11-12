import { useEffect, useRef, useState } from 'react';
import { getConfig } from '../../shared/api/config.js';
import { getCurrentUsername } from '../../shared/api/auth.js';
import { CmPenguinCompanion } from '../../shared/pet/CmPenguinCompanion.jsx';
import { getPetPageService, subscribePetPageService } from '../../shared/pet/pageService.js';

// Fullscreen AI-CUT does not mount UserLayout's pet. Reuse the same companion,
// including its chosen appearance, account isolation and page-service host.
export default function AICutPetHost({ scope }) {
  const username = getCurrentUsername();
  const [visible, setVisible] = useState(true);
  const [explicitlyOpened, setExplicitlyOpened] = useState(false);
  const consumedOpen = useRef({ scope: '', request: 0 });
  useEffect(() => {
    let active = true;
    setExplicitlyOpened(false);
    getConfig().then(config => { if (active) setVisible(config?.notifications?.petVisible !== false); }).catch(() => {});
    if (consumedOpen.current.scope !== scope) consumedOpen.current = { scope, request: 0 };
    const visibility = event => {
      setVisible(event.detail?.petVisible !== false); setExplicitlyOpened(false);
      const current = getPetPageService();
      if (current?.scope === scope) consumedOpen.current.request = current.openRequest;
    };
    const serviceOpened = () => {
      const service = getPetPageService();
      if (service?.scope === scope && service.openRequest > consumedOpen.current.request) {
        consumedOpen.current.request = service.openRequest;
        setExplicitlyOpened(true);
      }
    };
    const unsubscribe = subscribePetPageService(serviceOpened);
    serviceOpened();
    window.addEventListener('qiantie:notifications-updated', visibility);
    return () => { active = false; unsubscribe(); window.removeEventListener('qiantie:notifications-updated', visibility); };
  }, [username, scope]);
  if (!username || (!visible && !explicitlyOpened)) return null;
  return <CmPenguinCompanion username={username} accountSessionKey={username} />;
}
