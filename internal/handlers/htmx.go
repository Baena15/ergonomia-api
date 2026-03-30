package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// HXListProducts retorna HTML parcial para HTMX
func (h *Handler) HXListProducts(w http.ResponseWriter, r *http.Request) {
	// Obtener productos
	filters := models.ProductFilters{
		Category: r.URL.Query().Get("category"),
		Limit:    20,
	}

	result, err := h.store.ListProducts(r.Context(), filters)
	if err != nil {
		http.Error(w, "Error loading products", http.StatusInternalServerError)
		return
	}

	// Renderizar template simple (placeholder)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	tmpl := `
	<div class="products-grid">
		{{range .Products}}
		<article class="product-card" hx-get="/hx/products/{{.Slug}}" hx-target="#product-detail">
			<h3>{{.Name}}</h3>
			<p class="price">{{.Price}} {{.Currency}}</p>
			<p class="rating">⭐ {{.Rating}} ({{.Reviews}} reviews)</p>
			<p class="description">{{.Description}}</p>
			<button class="btn btn-primary" 
				hx-post="/api/v1/favorites" 
				hx-vals='{"product_id": {{.ID}}}'
				hx-swap="outerHTML">
				❤️ Guardar
			</button>
		</article>
		{{end}}
	</div>
	`
	t := template.Must(template.New("products").Parse(tmpl))
	t.Execute(w, result)
}

// HXProductDetail retorna detalle de producto para HTMX
func (h *Handler) HXProductDetail(w http.ResponseWriter, r *http.Request) {
	slug := getSlug(r)
	
	product, err := h.store.GetProductBySlug(r.Context(), slug)
	if err != nil || product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	fmt.Fprintf(w, `
	<div class="product-detail">
		<h2>%s</h2>
		<p class="price">%.2f %s</p>
		<p>%s</p>
		<div class="pros-cons">
			<div class="pros">
				<h4>Pros</h4>
				<ul>%s</ul>
			</div>
			<div class="cons">
				<h4>Contras</h4>
				<ul>%s</ul>
			</div>
		</div>
		<a href="%s" class="btn btn-primary" target="_blank" rel="nofollow sponsored">
			Ver en Amazon
		</a>
	</div>
	`, product.Name, product.Price, product.Currency, product.Description,
		formatList(product.Pros), formatList(product.Cons), product.AffiliateLinks["amazon"])
}

// HXCompareForm retorna el formulario de comparación
func (h *Handler) HXCompareForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
	<form hx-post="/hx/compare" hx-target="#comparison-result">
		<h3>Comparar productos</h3>
		<div class="compare-selects">
			<select name="product1" required>
				<option value="">Selecciona producto 1</option>
			</select>
			<select name="product2" required>
				<option value="">Selecciona producto 2</option>
			</select>
		</div>
		<button type="submit" class="btn btn-primary">Comparar</button>
	</form>
	<div id="comparison-result"></div>
	`)
}

// HXDoCompare realiza la comparación
func (h *Handler) HXDoCompare(w http.ResponseWriter, r *http.Request) {
	// Parsear form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	// Obtener IDs de productos
	p1 := r.FormValue("product1")
	p2 := r.FormValue("product2")

	// TODO: Implementar comparación real
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
	<div class="comparison-table">
		<h3>Comparación: %s vs %s</h3>
		<table>
			<tr><th>Característica</th><th>Producto 1</th><th>Producto 2</th></tr>
			<tr><td>Precio</td><td>...</td><td>...</td></tr>
		</table>
	</div>
	`, p1, p2)
}

// formatList formatea un slice de strings como lista HTML
func formatList(items []string) string {
	result := ""
	for _, item := range items {
		result += fmt.Sprintf("<li>%s</li>", item)
	}
	return result
}
