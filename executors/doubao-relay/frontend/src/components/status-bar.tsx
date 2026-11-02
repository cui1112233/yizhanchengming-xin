// 侧栏底部状态栏:"● 豆包"(绿已连 / 红未连 / 黄重连中)+ "● API"。
import { apiStatus, appConfig, appConfigLoading, appVersion, doubaoStatus } from '../state';
import { developerMode } from '../pages/settings';
import { apiStatusText, doubaoDot, doubaoStatusText, v88PairingStatus, versionLines } from '../view-model';
import type { DoubaoDot } from '../view-model';

function StatusDot({ color, label, detail }: { color: DoubaoDot; label: string; detail: string }) {
  return (
    <div class="status-item" data-color={color} title={detail}>
      <span class={`dot dot-${color}`} aria-hidden="true" />
      <span class="status-label">{label}</span>
      <span class="status-detail">{detail}</span>
    </div>
  );
}

export function StatusBar() {
  const s = doubaoStatus.value;
  const color: DoubaoDot = s ? doubaoDot(s) : 'red';
  const pairing = v88PairingStatus(appConfig.value, appConfigLoading.value);
  return (
    <div class="status-bar">
      <div class="status-title">状态</div>
      <StatusDot color={color} label="豆包" detail={doubaoStatusText(s)} />
      <StatusDot
        color={pairing.color}
        label="V88"
        detail={pairing.detail}
      />
      {developerMode.value && <StatusDot
        color={apiStatus.value?.listening ? 'green' : 'red'}
        label="API"
        detail={apiStatusText(apiStatus.value)}
      />}
      <div class="status-version">
        {versionLines(appVersion.value, s?.doubaoVersionDetected ?? null).map((line) => (
          <div key={line}>{line}</div>
        ))}
      </div>
    </div>
  );
}
