import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FaqAccordion } from './faq-accordion'

const items = [
  { q: 'Pertanyaan satu?', a: 'Jawaban satu.' },
  { q: 'Pertanyaan dua?', a: 'Jawaban dua.' },
]

describe('FaqAccordion', () => {
  it('renders all questions', () => {
    render(<FaqAccordion items={items} />)
    expect(screen.getByText('Pertanyaan satu?')).toBeInTheDocument()
    expect(screen.getByText('Pertanyaan dua?')).toBeInTheDocument()
  })

  it('opens the first item by default', () => {
    render(<FaqAccordion items={items} />)
    expect(screen.getByText('Jawaban satu.')).toBeInTheDocument()
  })

  it('toggles an item open and closed on click', async () => {
    const user = userEvent.setup()
    render(<FaqAccordion items={items} />)
    await user.click(screen.getByRole('button', { name: 'Pertanyaan dua?' }))
    expect(screen.getByText('Jawaban dua.')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Pertanyaan dua?' }))
    expect(screen.queryByText('Jawaban dua.')).not.toBeInTheDocument()
  })

  it('keeps only one item open at a time', async () => {
    const user = userEvent.setup()
    render(<FaqAccordion items={items} />)
    await user.click(screen.getByRole('button', { name: 'Pertanyaan dua?' }))
    expect(screen.getByText('Jawaban dua.')).toBeInTheDocument()
    expect(screen.queryByText('Jawaban satu.')).not.toBeInTheDocument()
  })

  it('wires aria-expanded and aria-controls', () => {
    render(<FaqAccordion items={items} />)
    const button = screen.getByRole('button', { name: 'Pertanyaan satu?' })
    expect(button).toHaveAttribute('aria-expanded', 'true')
    expect(button).toHaveAttribute('aria-controls', 'faq-panel-0')
  })

  it('renders the list variant without spacing wrapper classes', () => {
    const { container } = render(<FaqAccordion items={items} variant="list" />)
    expect(container.firstElementChild?.className).not.toContain('space-y-3')
  })
})
