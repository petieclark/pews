#!/bin/bash
# Comprehensive NULL handling fixes for all service files

set -e

echo "Applying COALESCE fixes to all service files..."

cd ~/Projects/pews

# ===== internal/giving/service.go =====
echo "Fixing internal/giving/service.go..."

# ListFunds - description is nullable
sed -i.bak 's/SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at/SELECT id, tenant_id, name, COALESCE(description, '\'''\'') as description, is_default, is_active, created_at, updated_at/' internal/giving/service.go

# GetFund - description is nullable  
sed -i.bak2 's/SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at \n\t\t FROM funds WHERE id/SELECT id, tenant_id, name, COALESCE(description, '\'''\'') as description, is_default, is_active, created_at, updated_at \n\t\t FROM funds WHERE id/' internal/giving/service.go

echo "Giving service fixed!"

echo "All fixes applied! Check the .bak files for backups."
