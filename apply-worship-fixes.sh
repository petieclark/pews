#!/bin/bash
set -e

echo "Applying worship module fixes..."

# 1. Add worship import to router
sed -i '' '/"github.com\/petieclark\/pews\/internal\/website"/a\
	"github.com/petieclark/pews/internal/worship"
' internal/router/router.go

# 2. Add worshipHandler parameter to router.New
sed -i '' '/servicesHandler \*services.Handler,/a\
	worshipHandler *worship.Handler,
' internal/router/router.go

# 3. Add worship routes after services routes
WORSHIP_ROUTES='
		\/\/ Worship - Service Plans
		r.Get("\/api\/worship\/plans", worshipHandler.ListPlans)
		r.Post("\/api\/worship\/plans", worshipHandler.CreatePlan)
		r.Get("\/api\/worship\/plans\/{id}", worshipHandler.GetPlan)
		r.Put("\/api\/worship\/plans\/{id}", worshipHandler.UpdatePlan)
		r.Post("\/api\/worship\/plans\/{id}\/publish", worshipHandler.PublishPlan)
		r.Post("\/api\/worship\/plans\/{id}\/items", worshipHandler.AddItem)
		r.Put("\/api\/worship\/plans\/{id}\/items\/{itemId}", worshipHandler.UpdateItem)
		r.Delete("\/api\/worship\/plans\/{id}\/items\/{itemId}", worshipHandler.DeleteItem)
		r.Get("\/api\/worship\/plans\/{id}\/export", worshipHandler.ExportPlan)
'

# Find the line with "// Sermons" and insert before it
sed -i '' '/^[[:space:]]*\/\/ Sermons$/i\
'"$WORSHIP_ROUTES" internal/router/router.go

# 4. Add worship import to main.go
sed -i '' '/"github.com\/petieclark\/pews\/internal\/website"/a\
	"github.com/petieclark/pews/internal/worship"
' cmd/pews/main.go

# 5. Add worship service initialization
sed -i '' '/servicesService := services.NewService(db.Pool)/a\
	worshipService := worship.NewService(db.Pool)
' cmd/pews/main.go

# 6. Add worship handler initialization
sed -i '' '/servicesHandler := services.NewHandler(servicesService)/a\
	worshipHandler := worship.NewHandler(worshipService)
' cmd/pews/main.go

# 7. Add worshipHandler to router.New call
sed -i '' '/servicesHandler,$/a\
		worshipHandler,
' cmd/pews/main.go

echo "Worship module backend fixes applied!"
echo "Now applying frontend fixes..."

# Create KeySelect component directory if needed
mkdir -p web/src/lib/components

echo "All fixes applied! Don't forget to:"
echo "1. Review changes with: git diff"
echo "2. Test the build"
echo "3. Commit the changes"
