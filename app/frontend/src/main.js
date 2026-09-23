import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import HomeView from './views/HomeView.vue'
import ProductView from './views/ProductView.vue'
import './styles.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/products/:id', name: 'product', component: ProductView, props: true },
  ],
})

createApp(App).use(router).mount('#app')