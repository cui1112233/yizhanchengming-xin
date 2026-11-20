import { Button, Empty } from 'antd'
import { Plus } from 'lucide-react'

import { createHomeModel } from './homeModel.js'

export function HomePage() {
  const copy = createHomeModel()

  return (
    <main className="magic-home">
      <header className="magic-home__header">
        <div>
          <p className="magic-home__eyebrow">MAGIC CUT WORKSPACE</p>
          <h1>{copy.sectionTitle}</h1>
        </div>
        <Button type="primary" size="large" icon={<Plus size={17} />}>
          {copy.primaryAction}
        </Button>
      </header>
      <section className="magic-home__empty" aria-label={copy.sectionTitle}>
        <Empty description={copy.emptyTitle} />
        <p>{copy.emptyDescription}</p>
      </section>
    </main>
  )
}
