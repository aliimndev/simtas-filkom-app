'use client'

import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}

export default function UiPreviewPage() {
  return (
    <div className="mx-auto max-w-4xl space-y-6 py-10">
      <div className="text-center">
        <h1 className="text-3xl font-bold">UI Preview</h1>
        <p className="mt-1 text-sm text-muted-foreground">Preview komponen design system SIMTAS (hanya development)</p>
      </div>

      <Section title="Button">
        <div className="flex flex-wrap gap-2">
          <Button variant="primary">Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="success">Success</Button>
          <Button variant="danger">Danger</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="outline">Outline</Button>
          <Button loading>Loading</Button>
          <Button disabled>Disabled</Button>
          <Button size="sm">Small</Button>
          <Button size="lg">Large</Button>
        </div>
      </Section>

      <Section title="Badge">
        <div className="flex flex-wrap gap-2">
          <Badge>Default</Badge>
          <Badge variant="primary">Primary</Badge>
          <Badge variant="success">Success</Badge>
          <Badge variant="warning">Warning</Badge>
          <Badge variant="danger">Danger</Badge>
          <Badge variant="muted">Muted</Badge>
          <Badge variant="secondary">Secondary</Badge>
        </div>
      </Section>

      <Section title="Alert">
        <div className="space-y-2">
          <Alert variant="info">Info: Informasi umum.</Alert>
          <Alert variant="success">Sukses: Operasi berhasil.</Alert>
          <Alert variant="warning">Peringatan: Perhatikan ini.</Alert>
          <Alert variant="danger">Error: Terjadi kesalahan.</Alert>
        </div>
      </Section>

      <Section title="Form">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="preview-input">Input Text</Label>
            <Input id="preview-input" placeholder="Placeholder" />
          </div>
          <div>
            <Label htmlFor="preview-input-invalid">Input Invalid</Label>
            <Input id="preview-input-invalid" placeholder="Invalid state" invalid />
          </div>
          <div>
            <Label htmlFor="preview-select">Select</Label>
            <Select id="preview-select">
              <option>Pilihan 1</option>
              <option>Pilihan 2</option>
            </Select>
          </div>
          <div>
            <Label htmlFor="preview-textarea">Textarea</Label>
            <Textarea id="preview-textarea" placeholder="Text area…" />
          </div>
        </div>
      </Section>

      <Section title="Table">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Nama</TableHead>
              <TableHead>Peran</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell className="font-medium">Mahasiswa A</TableCell>
              <TableCell>Mahasiswa</TableCell>
              <TableCell><Badge variant="success">Aktif</Badge></TableCell>
            </TableRow>
            <TableRow>
              <TableCell className="font-medium">Dosen B</TableCell>
              <TableCell>Dosen Pembimbing</TableCell>
              <TableCell><Badge variant="muted">Nonaktif</Badge></TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Section>

      <Section title="Loading States">
        <div className="space-y-4">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-20 w-full" />
          <Spinner label="Memuat contoh…" />
        </div>
      </Section>
    </div>
  )
}
