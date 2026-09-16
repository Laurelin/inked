# inked

Tailwind-style HTML templates to PDF, in pure Go.

Inked owns its layout and PDF serialization pipeline. It does not invoke a
browser or depend on a third-party PDF generator.

## Usage

```go
renderer, err := inked.New()
if err != nil {
    return err
}

if err := renderer.Render(output, `<p class="text-xl">Hello</p>`); err != nil {
    return err
}
```

## Supported template subset

The initial layout engine supports `body`, `div`, `p`, `span`, and `br`.
`head`, `script`, and `style` content is ignored. Other elements return an
explicit error.

Supported utility classes:

- Display: `block`, `inline`
- Font size: `text-xs`, `text-sm`, `text-base`, `text-lg`, `text-xl`
- Font weight: `font-normal`, `font-bold`
- Line height: `leading-none`, `leading-tight`, `leading-normal`,
  `leading-relaxed`
- Alignment: `text-left`, `text-center`, `text-right`
- Spacing: `m`, `mx`, `my`, `mt`, `mr`, `mb`, `ml`, and the corresponding
  `p` utilities, with scales `0`, `1`, `2`, `4`, `6`, and `8`

Spacing follows a Tailwind-style four-pixel scale converted to PDF points:
one spacing unit is three points. Unknown classes return an explicit error.

Documents use US Letter pages, 72-point page insets, and the PDF standard
Helvetica and Helvetica-Bold fonts. Text must be representable in WinAnsi;
unsupported characters return an explicit error.
