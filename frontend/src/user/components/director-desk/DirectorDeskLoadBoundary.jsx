import { Component } from 'react';

export class DirectorDeskLoadBoundary extends Component {
  state = { error: null };

  static getDerivedStateFromError(error) {
    return { error };
  }

  render() {
    if (!this.state.error) return this.props.children;
    return <section className="agent-director-desk-load-error" role="alert">
      <strong>导演台资源加载失败</strong>
      <span>本地运行包可能仍在使用旧资源，请重新加载当前页面。</span>
      <button type="button" onClick={() => (this.props.onRetry || (() => window.location.reload()))()}>重新加载页面</button>
    </section>;
  }
}
