import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/Home'
import Observer from '@/components/Observer'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/:roomName',
      name: 'Observer',
      component: Observer
    }
  ]
})
