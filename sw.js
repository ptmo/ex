// Instalasi Service Worker
self.addEventListener('install', (e) => {
    console.log('[Service Worker] Terinstal');
    self.skipWaiting();
});

// Aktivasi Service Worker
self.addEventListener('activate', (e) => {
    console.log('[Service Worker] Aktif');
});

// Mencegat Fetch Request (Syarat wajib PWA Chrome)
self.addEventListener('fetch', (e) => {
    // Biarkan aplikasi mengambil data dari internet seperti biasa
    e.respondWith(fetch(e.request).catch(() => {
        return new Response('Anda sedang offline.');
    }));
});
