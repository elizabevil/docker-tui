import {createApp} from 'vue'
import {createRouter, createWebHashHistory} from 'vue-router'
import App from './App.vue'
import ContainersPage from './pages/ContainersPage.vue'
import ImagesPage from './pages/ImagesPage.vue'
import VolumesPage from './pages/VolumesPage.vue'
import NetworksPage from './pages/NetworksPage.vue'
import DetailPage from './pages/DetailPage.vue'
import LogViewPage from './pages/LogViewPage.vue'
import HelpPage from './pages/HelpPage.vue'
import CatalogPage from './pages/CatalogPage.vue'
import './styles/terminal.css'
import './styles/themes/nord.css'
import './styles/themes/dracula.css'

const routes = [
    {path: '/', redirect: '/containers'},
    {path: '/containers', component: ContainersPage, meta: {title: 'Containers'}},
    {path: '/images', component: ImagesPage, meta: {title: 'Images'}},
    {path: '/volumes', component: VolumesPage, meta: {title: 'Volumes'}},
    {path: '/networks', component: NetworksPage, meta: {title: 'Networks'}},
    {path: '/detail', component: DetailPage, meta: {title: 'Detail'}},
    {path: '/logs', component: LogViewPage, meta: {title: 'Logs'}},
    {path: '/help', component: HelpPage, meta: {title: 'Help'}},
    {path: '/catalog', component: CatalogPage, meta: {title: 'Catalog'}},
]

const router = createRouter({history: createWebHashHistory(), routes})
createApp(App).use(router).mount('#app')
