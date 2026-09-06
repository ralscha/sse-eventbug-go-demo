import App from './app';

const app = new App();
app.start();
window.addEventListener('pagehide', () => {
    app.stop();
    window.removeEventListener('resize', app.onResize);
}, { once: true });
