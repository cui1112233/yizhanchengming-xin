import { useRef } from 'react';

export function Link({ href, children, className, onClick, onIntent, reload = false }) {
  const didIntent = useRef(false);
  const triggerIntent = () => {
    if (didIntent.current) return;
    didIntent.current = true;
    onIntent?.();
  };

  return (
    <a
      href={href}
      className={className}
      onMouseEnter={triggerIntent}
      onFocus={triggerIntent}
      onClick={(event) => {
        onClick?.(event);
        if (event.defaultPrevented) return;
        if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
        if (reload) return;
        event.preventDefault();
        window.history.pushState({}, '', href);
        window.dispatchEvent(new PopStateEvent('popstate'));
      }}
    >
      {children}
    </a>
  );
}
