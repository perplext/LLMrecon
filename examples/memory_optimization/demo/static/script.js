
// JavaScript for the static file handler example
document.addEventListener('DOMContentLoaded', function() {
	const memoryStatsElement = document.getElementById('memory-stats');
	const monitoringStatsElement = document.getElementById('monitoring-stats');
	
	function updateStats() {
		fetch('/stats')
			.then(response => response.json())
			.then(data => {
				// Update memory stats
				const memoryData = {
					heapAlloc: data.heapAlloc + ' MB',
					heapObjects: data.heapObjects,
					gcCPUFraction: data.gcCPUFraction
				};
				memoryStatsElement.innerHTML = JSON.stringify(memoryData, null, 2);
				
				// Update monitoring stats
				const monitoringData = {
					filesServed: data.filesServed,
					cacheHits: data.cacheHits,
					cacheMisses: data.cacheMisses,
					cacheHitRatio: (data.monitoring.cacheHitRatio * 100).toFixed(2) + '%',
					averageServeTime: data.monitoring.averageServeTimeMs + ' ms'
				};
				monitoringStatsElement.innerHTML = JSON.stringify(monitoringData, null, 2);
			})
			.catch(error => {
				console.error('Error fetching stats:', error);
			});
	}
	
	// Update stats initially and then every 2 seconds
	updateStats();
	setInterval(updateStats, 2000);
});
