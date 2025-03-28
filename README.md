# backend-petdoc-go

Pengembangan Backend PetDoc menggunakan Go.

## Tentang Pointer Receiver

Dalam Go, receiver pada method dapat berupa value atau pointer. Berikut penjelasannya:

### Contoh Kode

```go
type Circle struct {
    radius float64
}

// Metode dengan value receiver
func (c Circle) area() float64 {
    return 3.14 * c.radius * c.radius
}

// Metode dengan pointer receiver
func (c *Circle) setRadius(radius float64) {
    c.radius = radius
}

Ada dua jenis utama receiver:

Value Receiver:
Receiver adalah salinan dari instance tipe.
Perubahan yang dilakukan pada receiver dalam metode tidak memengaruhi instance asli.
Ditandai dengan receiver yang tidak menggunakan pointer.
Contoh: func (t Tipe) NamaMetode() {}
Pointer Receiver:
Receiver adalah pointer ke instance tipe.
Perubahan yang dilakukan pada receiver dalam metode memengaruhi instance asli.
Ditandai dengan receiver yang menggunakan pointer.
Contoh: func (t *Tipe) NamaMetode() {}

## kelebihan pointer gak perlu return nilainya langsung berubah