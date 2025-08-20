#!/bin/bash

echo "Fixing Go syntax errors in analytics and API packages..."

# Fix missing closing braces in collector.go
echo "Fixing src/analytics/collector.go..."
sed -i.bak '124a\
}' src/analytics/collector.go

sed -i.bak '/^type NetworkStats struct {/,/^$/ {
    /^$/i\
}
}' src/analytics/collector.go

# Fix missing closing braces in dashboard.go structs
echo "Fixing src/analytics/dashboard.go..."
sed -i.bak '/^type Dashboard struct {/,/^type/ {
    /^type Layout struct {/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type Layout struct {/,/^type/ {
    /^type LayoutType/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type GridSize struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type BreakpointConfig struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type DashboardWidget struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type WidgetPosition struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type WidgetDataSource struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type DashboardFilter struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

sed -i.bak '/^type DashboardPermission struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/dashboard.go

# Fix executive.go
echo "Fixing src/analytics/executive.go..."
sed -i.bak '/^type ExecutiveDashboard struct {/,/^type/ {
    /^type ExecutiveMetrics/i\
}
}' src/analytics/executive.go

sed -i.bak '/^type ExecutiveMetrics struct {/,/^func/ {
    /^func/i\
}
}' src/analytics/executive.go

# Fix types.go structs
echo "Fixing src/analytics/types.go..."
for struct in "Metric" "AggregatedMetric" "Alert" "AlertRule" "TimeWindow" "DataPoint" "Visualization" "DashboardConfig" "ExportOptions" "ReportConfig" "AnalyticsConfig"; do
    sed -i.bak "/^type $struct struct {/,/^type/ {
        /^type/i\\
}
    }" src/analytics/types.go
done

# Fix export.go
echo "Fixing src/analytics/export.go..."
sed -i.bak '/^type ExportHandler struct {/,/^type/ {
    /^type/i\
}
}' src/analytics/export.go

# Fix comparative.go function
echo "Fixing src/analytics/comparative.go..."
sed -i.bak '/^func.*Compare.*{$/,/^func/ {
    /^func [^{]*$/i\
}
}' src/analytics/comparative.go

# Fix API handlers.go functions
echo "Fixing src/api/handlers.go..."
sed -i.bak 's/^func handleVersion/func (s *Server) handleVersion/' src/api/handlers.go
sed -i.bak 's/^func handleCreateScan/func (s *Server) handleCreateScan/' src/api/handlers.go
sed -i.bak 's/^func handleListScans/func (s *Server) handleListScans/' src/api/handlers.go
sed -i.bak 's/^func handleGetScan/func (s *Server) handleGetScan/' src/api/handlers.go
sed -i.bak 's/^func handleCancelScan/func (s *Server) handleCancelScan/' src/api/handlers.go
sed -i.bak 's/^func handleGetScanResults/func (s *Server) handleGetScanResults/' src/api/handlers.go
sed -i.bak 's/^func handleListTemplates/func (s *Server) handleListTemplates/' src/api/handlers.go
sed -i.bak 's/^func handleGetTemplate/func (s *Server) handleGetTemplate/' src/api/handlers.go
sed -i.bak 's/^func handleListCategories/func (s *Server) handleListCategories/' src/api/handlers.go
sed -i.bak 's/^func handleListModules/func (s *Server) handleListModules/' src/api/handlers.go
sed -i.bak 's/^func handleGetModule/func (s *Server) handleGetModule/' src/api/handlers.go

# Fix auth_handlers.go functions
echo "Fixing src/api/auth_handlers.go..."
sed -i.bak 's/^func handleRefreshToken/func (s *Server) handleRefreshToken/' src/api/auth_handlers.go
sed -i.bak 's/^func handleCreateUser/func (s *Server) handleCreateUser/' src/api/auth_handlers.go
sed -i.bak 's/^func handleUpdatePassword/func (s *Server) handleUpdatePassword/' src/api/auth_handlers.go
sed -i.bak 's/^func handleCreateAPIKey/func (s *Server) handleCreateAPIKey/' src/api/auth_handlers.go
sed -i.bak 's/^func handleListAPIKeys/func (s *Server) handleListAPIKeys/' src/api/auth_handlers.go
sed -i.bak 's/^func handleGetAPIKey/func (s *Server) handleGetAPIKey/' src/api/auth_handlers.go
sed -i.bak 's/^func handleRevokeAPIKey/func (s *Server) handleRevokeAPIKey/' src/api/auth_handlers.go
sed -i.bak 's/^func jwtMiddleware/func (s *Server) jwtMiddleware/' src/api/auth_handlers.go
sed -i.bak 's/^func handleGetProfile/func (s *Server) handleGetProfile/' src/api/auth_handlers.go

# Fix middleware.go functions
echo "Fixing src/api/middleware.go..."
sed -i.bak 's/^func corsMiddleware/func (s *Server) corsMiddleware/' src/api/middleware.go
sed -i.bak 's/^func jsonContentTypeMiddleware/func (s *Server) jsonContentTypeMiddleware/' src/api/middleware.go
sed -i.bak 's/^func authMiddleware/func (s *Server) authMiddleware/' src/api/middleware.go
sed -i.bak 's/^func rateLimitMiddleware/func (s *Server) rateLimitMiddleware/' src/api/middleware.go
sed -i.bak 's/^func extractAPIKey/func (s *Server) extractAPIKey/' src/api/middleware.go
sed -i.bak 's/^func isValidAPIKey/func (s *Server) isValidAPIKey/' src/api/middleware.go
sed -i.bak 's/^func writeJSON/func (s *Server) writeJSON/' src/api/middleware.go

# Fix security_middleware.go functions
echo "Fixing src/api/security_middleware.go..."
sed -i.bak 's/^func securityHeadersMiddleware/func (s *Server) securityHeadersMiddleware/' src/api/security_middleware.go
sed -i.bak 's/^func ipWhitelistMiddleware/func (s *Server) ipWhitelistMiddleware/' src/api/security_middleware.go
sed -i.bak 's/^func requestSizeLimitMiddleware/func (s *Server) requestSizeLimitMiddleware/' src/api/security_middleware.go

# Remove backup files
find src -name "*.bak" -delete

echo "Syntax fixes applied. Please run 'go fmt ./...' to format the code."