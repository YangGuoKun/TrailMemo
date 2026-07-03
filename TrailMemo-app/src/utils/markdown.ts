export function renderMarkdownToRichText(markdown: string): string {
  const normalized = markdown.replace(/\r\n/g, '\n').trim()
  if (!normalized) return ''

  const lines = normalized.split('\n')
  const blocks: string[] = []

  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i]

    if (!line.trim()) continue

    if (line.trim().startsWith('```')) {
      const codeLines: string[] = []
      i += 1
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i])
        i += 1
      }
      blocks.push(`<pre class="md-pre"><code class="md-code">${escapeHtml(codeLines.join('\n'))}</code></pre>`)
      continue
    }

    const headingMatch = line.match(/^(#{1,3})\s+(.+)$/)
    if (headingMatch) {
      const level = headingMatch[1].length
      blocks.push(`<h${level} class="md-h${level}">${renderInline(headingMatch[2])}</h${level}>`)
      continue
    }

    if (/^\s*[-*]\s+/.test(line)) {
      const items: string[] = []
      while (i < lines.length && /^\s*[-*]\s+/.test(lines[i])) {
        items.push(`<li class="md-li">${renderInline(lines[i].replace(/^\s*[-*]\s+/, ''))}</li>`)
        i += 1
      }
      i -= 1
      blocks.push(`<ul class="md-ul">${items.join('')}</ul>`)
      continue
    }

    if (/^\s*\d+\.\s+/.test(line)) {
      const items: string[] = []
      while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
        items.push(`<li class="md-li">${renderInline(lines[i].replace(/^\s*\d+\.\s+/, ''))}</li>`)
        i += 1
      }
      i -= 1
      blocks.push(`<ol class="md-ol">${items.join('')}</ol>`)
      continue
    }

    if (/^\s*>\s?/.test(line)) {
      const quoteLines: string[] = []
      while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
        quoteLines.push(lines[i].replace(/^\s*>\s?/, ''))
        i += 1
      }
      i -= 1
      blocks.push(`<blockquote class="md-blockquote">${quoteLines.map(renderInline).join('<br/>')}</blockquote>`)
      continue
    }

    const paragraphLines = [line]
    while (
      i + 1 < lines.length &&
      lines[i + 1].trim() &&
      !lines[i + 1].trim().startsWith('```') &&
      !/^(#{1,3})\s+/.test(lines[i + 1]) &&
      !/^\s*[-*]\s+/.test(lines[i + 1]) &&
      !/^\s*\d+\.\s+/.test(lines[i + 1]) &&
      !/^\s*>\s?/.test(lines[i + 1])
    ) {
      i += 1
      paragraphLines.push(lines[i])
    }

    blocks.push(`<p class="md-p">${paragraphLines.map(renderInline).join('<br/>')}</p>`)
  }

  return blocks.join('')
}

function renderInline(text: string): string {
  return escapeHtml(text)
    .replace(/`([^`]+)`/g, '<code class="md-code">$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong class="md-strong">$1</strong>')
    .replace(/__([^_]+)__/g, '<strong class="md-strong">$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em class="md-em">$1</em>')
    .replace(/_([^_]+)_/g, '<em class="md-em">$1</em>')
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
