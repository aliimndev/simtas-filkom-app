import { RedirectAuthenticated } from '@/components/features/landing/redirect-authenticated'
import { HeroSection } from '@/components/features/landing/hero'
import { ProblemSolutionSection } from '@/components/features/landing/stats'
import { PipelineSection } from '@/components/features/landing/pipeline'
import { FeaturesSection } from '@/components/features/landing/features'
import { CtaSection } from '@/components/features/landing/cta'

export const metadata = {
  title: 'SIMTAS FILKOM',
  description:
    'Satu sistem untuk seluruh perjalanan Tugas Akhir Skripsi — pengajuan judul, bimbingan, seminar, sidang, dan arsip. SIMTAS FILKOM Universitas Djuanda.',
}

export default function HomePage() {
  return (
    <div>
      <RedirectAuthenticated />
      <HeroSection />
      <ProblemSolutionSection />
      <PipelineSection />
      <FeaturesSection />
      <CtaSection />
    </div>
  )
}
