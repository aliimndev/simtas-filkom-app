import { render, screen } from '@testing-library/react'
import { Badge } from './badge'

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge>Approved</Badge>)
    expect(screen.getByText('Approved')).toBeInTheDocument()
  })

  it('applies the default variant class', () => {
    render(<Badge>Default</Badge>)
    const badge = screen.getByText('Default')
    expect(badge.className).toContain('bg-primary')
  })

  it('applies the success variant class', () => {
    render(<Badge variant="success">Ok</Badge>)
    expect(screen.getByText('Ok').className).toContain('bg-success/15')
  })

  it('applies the danger variant class', () => {
    render(<Badge variant="danger">Error</Badge>)
    expect(screen.getByText('Error').className).toContain('bg-danger/15')
  })

  it('merges custom className', () => {
    render(<Badge className="custom-class">Mix</Badge>)
    expect(screen.getByText('Mix').className).toContain('custom-class')
  })

  it('forwards extra props (aria-label)', () => {
    render(<Badge aria-label="status">X</Badge>)
    expect(screen.getByLabelText('status')).toBeInTheDocument()
  })
})
