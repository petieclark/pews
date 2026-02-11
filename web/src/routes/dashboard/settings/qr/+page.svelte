<script lang="ts">
	import { onMount } from 'svelte';

	let tenantSubdomain = '';
	let customUrl = '';
	let customSize = 300;
	let selectedStation = '';
	let stations: any[] = [];

	const defaultSize = 300;

	onMount(async () => {
		// Extract tenant subdomain from hostname
		const hostname = window.location.hostname;
		const parts = hostname.split('.');
		if (parts.length > 2) {
			tenantSubdomain = parts[0];
		} else {
			tenantSubdomain = 'demo'; // fallback
		}

		// Load stations for check-in QR
		await loadStations();
	});

	async function loadStations() {
		try {
			const response = await fetch('/api/checkins/stations', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				stations = await response.json();
				if (stations.length > 0) {
					selectedStation = stations[0].id;
				}
			}
		} catch (error) {
			console.error('Failed to load stations:', error);
		}
	}

	function getQRUrl(type: string, size: number = defaultSize): string {
		const baseUrl = '/api/qr';
		switch (type) {
			case 'checkin':
				return `${baseUrl}/checkin?station=${selectedStation}&size=${size}`;
			case 'connect':
				return `${baseUrl}/connect?size=${size}`;
			case 'give':
				return `${baseUrl}/give?size=${size}`;
			case 'prayer':
				return `${baseUrl}/prayer?size=${size}`;
			case 'custom':
				return `${baseUrl}/custom?url=${encodeURIComponent(customUrl)}&size=${customSize}`;
			default:
				return '';
		}
	}

	function downloadQR(type: string, filename: string) {
		const url = getQRUrl(type);
		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}

	function printQR(type: string, title: string, instructions: string) {
		const qrUrl = getQRUrl(type);
		
		const printWindow = window.open('', '_blank');
		if (!printWindow) return;

		printWindow.document.write(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Print QR Code - ${title}</title>
				<style>
					@media print {
						@page { margin: 1in; }
					}
					body {
						font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
						text-align: center;
						padding: 40px 20px;
					}
					h1 {
						color: #1B3A4B;
						font-size: 28px;
						margin-bottom: 10px;
					}
					.instructions {
						color: #666;
						font-size: 18px;
						margin-bottom: 30px;
					}
					img {
						max-width: 400px;
						margin: 30px auto;
						display: block;
					}
					.footer {
						margin-top: 40px;
						color: #999;
						font-size: 14px;
					}
				</style>
			</head>
			<body>
				<h1>${title}</h1>
				<div class="instructions">${instructions}</div>
				<img src="${qrUrl}" alt="${title}" />
				<div class="footer">Powered by Pews</div>
			</body>
			</html>
		`);

		printWindow.document.close();
		
		// Wait for image to load before printing
		const img = printWindow.document.querySelector('img');
		if (img) {
			img.onload = () => {
				setTimeout(() => {
					printWindow.print();
				}, 500);
			};
		}
	}

	function generateCustomQR() {
		if (!customUrl) {
			alert('Please enter a URL');
			return;
		}
		// The QR will be displayed via the reactive getQRUrl call
	}
</script>

<div class="max-w-6xl">
	<a href="/dashboard/settings" class="text-sm text-secondary hover:text-[var(--teal)] mb-4 inline-block">← Back to Settings</a>
	<div class="mb-6">
		<h1 class="text-2xl sm:text-3xl font-bold text-primary">QR Codes</h1>
		<p class="text-secondary mt-1">Generate printable QR codes for check-ins, connection cards, and more</p>
	</div>

	<!-- Standard QR Codes Grid -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
		<!-- Check-in QR -->
		<div class="bg-surface rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-primary mb-4">Check-In Station</h2>
			
			{#if stations.length > 0}
				<div class="mb-4">
					<label class="block text-sm font-medium text-primary mb-2">Station</label>
					<select 
						bind:value={selectedStation}
						class="w-full px-3 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
					>
						{#each stations as station}
							<option value={station.id}>{station.name}</option>
						{/each}
					</select>
				</div>

				<div class="bg-[var(--surface-hover)] rounded-lg p-4 mb-4 flex justify-center">
					<img 
						src={getQRUrl('checkin')} 
						alt="Check-in QR Code" 
						class="w-48 h-48"
					/>
				</div>

				<div class="flex gap-2">
					<button
						on:click={() => downloadQR('checkin', 'checkin-qr.png')}
						class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
					>
						Download
					</button>
					<button
						on:click={() => printQR('checkin', 'Check-In', 'Scan to check in to this service')}
						class="flex-1 px-4 py-2 border-2 border-[#4A8B8C] text-[#4A8B8C] font-semibold rounded-lg hover:bg-[#4A8B8C] hover:text-white transition"
					>
						Print
					</button>
				</div>
			{:else}
				<p class="text-secondary text-sm mb-4">No check-in stations found. Create one first.</p>
				<a 
					href="/dashboard/checkins/stations"
					class="inline-block px-4 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
				>
					Create Station
				</a>
			{/if}
		</div>

		<!-- Connection Card QR -->
		<div class="bg-surface rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-primary mb-4">Connection Card</h2>
			
			<div class="bg-[var(--surface-hover)] rounded-lg p-4 mb-4 flex justify-center">
				<img 
					src={getQRUrl('connect')} 
					alt="Connection Card QR Code" 
					class="w-48 h-48"
				/>
			</div>

			<div class="flex gap-2">
				<button
					on:click={() => downloadQR('connect', 'connect-qr.png')}
					class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
				>
					Download
				</button>
				<button
					on:click={() => printQR('connect', 'Connect With Us', 'Scan to fill out a connection card')}
					class="flex-1 px-4 py-2 border-2 border-[#4A8B8C] text-[#4A8B8C] font-semibold rounded-lg hover:bg-[#4A8B8C] hover:text-white transition"
				>
					Print
				</button>
			</div>

			<p class="text-xs text-secondary mt-3">Perfect for pew cards and seat backs</p>
		</div>

		<!-- Giving QR -->
		<div class="bg-surface rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-primary mb-4">Give</h2>
			
			<div class="bg-[var(--surface-hover)] rounded-lg p-4 mb-4 flex justify-center">
				<img 
					src={getQRUrl('give')} 
					alt="Give QR Code" 
					class="w-48 h-48"
				/>
			</div>

			<div class="flex gap-2">
				<button
					on:click={() => downloadQR('give', 'give-qr.png')}
					class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
				>
					Download
				</button>
				<button
					on:click={() => printQR('give', 'Give Online', 'Scan to give securely')}
					class="flex-1 px-4 py-2 border-2 border-[#4A8B8C] text-[#4A8B8C] font-semibold rounded-lg hover:bg-[#4A8B8C] hover:text-white transition"
				>
					Print
				</button>
			</div>

			<p class="text-xs text-secondary mt-3">Display during services or in your bulletin</p>
		</div>

		<!-- Prayer QR -->
		<div class="bg-surface rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-primary mb-4">Prayer Requests</h2>
			
			<div class="bg-[var(--surface-hover)] rounded-lg p-4 mb-4 flex justify-center">
				<img 
					src={getQRUrl('prayer')} 
					alt="Prayer QR Code" 
					class="w-48 h-48"
				/>
			</div>

			<div class="flex gap-2">
				<button
					on:click={() => downloadQR('prayer', 'prayer-qr.png')}
					class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
				>
					Download
				</button>
				<button
					on:click={() => printQR('prayer', 'Prayer Requests', 'Scan to submit a prayer request')}
					class="flex-1 px-4 py-2 border-2 border-[#4A8B8C] text-[#4A8B8C] font-semibold rounded-lg hover:bg-[#4A8B8C] hover:text-white transition"
				>
					Print
				</button>
			</div>

			<p class="text-xs text-secondary mt-3">Easy way to collect prayer requests</p>
		</div>
	</div>

	<!-- Custom QR Code Builder -->
	<div class="bg-surface rounded-lg shadow-lg p-6">
		<h2 class="text-xl font-bold text-primary mb-4">Custom QR Code</h2>
		<p class="text-secondary mb-4">Generate a QR code for any URL</p>

		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
			<div class="md:col-span-2">
				<label class="block text-sm font-medium text-primary mb-2">URL</label>
				<input
					type="url"
					bind:value={customUrl}
					placeholder="https://example.com"
					class="w-full px-3 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-primary mb-2">Size (px)</label>
				<input
					type="number"
					bind:value={customSize}
					min="100"
					max="1000"
					step="50"
					class="w-full px-3 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				/>
			</div>
		</div>

		<button
			on:click={generateCustomQR}
			disabled={!customUrl}
			class="px-6 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition disabled:opacity-50 disabled:cursor-not-allowed mb-4"
		>
			Generate QR Code
		</button>

		{#if customUrl}
			<div class="mt-4 flex flex-col items-center">
				<div class="bg-[var(--surface-hover)] rounded-lg p-4 mb-4">
					<img 
						src={getQRUrl('custom', customSize)} 
						alt="Custom QR Code" 
						style="width: {customSize}px; height: {customSize}px;"
					/>
				</div>

				<button
					on:click={() => {
						const url = getQRUrl('custom', customSize);
						const link = document.createElement('a');
						link.href = url;
						link.download = 'custom-qr.png';
						document.body.appendChild(link);
						link.click();
						document.body.removeChild(link);
					}}
					class="px-6 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition"
				>
					Download Custom QR
				</button>
			</div>
		{/if}
	</div>

	<!-- Usage Tips -->
	<div class="mt-8 p-6 bg-[var(--surface-hover)] border border-custom rounded-lg">
		<h3 class="font-bold text-primary mb-2">💡 Tips for Using QR Codes</h3>
		<ul class="list-disc list-inside text-sm text-secondary space-y-1">
			<li><strong>Check-in:</strong> Print and place at entrance tables for fast self-service check-in</li>
			<li><strong>Connection cards:</strong> Perfect for pew backs, seat cards, or bulletin inserts</li>
			<li><strong>Giving:</strong> Display on screens during services or in printed materials</li>
			<li><strong>Prayer:</strong> Place near prayer walls or distribute in small groups</li>
			<li><strong>Size matters:</strong> Larger codes (400-500px) scan better from a distance</li>
			<li><strong>Test first:</strong> Always test QR codes with your phone camera before mass printing</li>
		</ul>
	</div>
</div>

<style>
	/* Ensure QR images load properly */
	img {
		image-rendering: -webkit-optimize-contrast;
		image-rendering: crisp-edges;
	}
</style>
