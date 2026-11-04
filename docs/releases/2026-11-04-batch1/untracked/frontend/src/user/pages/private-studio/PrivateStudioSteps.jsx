/** 左栏三段进度，文案与视觉取自设计规格图。 */
export function PrivateStudioSteps({ steps = [] }) {
  return <div className="private-studio-steps" aria-label="创作进度">
    {steps.map((step, index) => <div key={step.key || index} className={`private-studio-step is-${step.state || 'pending'}`}>
      <span className="private-studio-step-dot" aria-hidden="true">{step.state === 'done' ? '✓' : index + 1}</span>
      <span className="private-studio-step-label">{step.label}</span>
    </div>)}
  </div>;
}

export default PrivateStudioSteps;
