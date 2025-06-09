class Footer {
    constructor() {
        this.init();
    }

    init() {
        this.updateYear();
    }

    updateYear() {
        const yearElement = document.querySelector('.footer-bottom p');
        if (yearElement) {
            const currentYear = new Date().getFullYear();
            yearElement.innerHTML = yearElement.innerHTML.replace('2024', currentYear);
        }
    }
}

export default Footer; 