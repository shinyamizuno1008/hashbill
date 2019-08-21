import Vue from "vue";
import VueRouter from "vue-router";

import RegisterEvent from "./components/RegisterEvent";
import EventList from "./components/EventList";

Vue.use(VueRouter);

export default new VueRouter({
  mode: "history",
  routes: [
    {
      path: "/",
      name: "home",
      component: RegisterEvent
    },
    {
      path: "/eventlist",
      name: "event list",
      component: EventList
    }
  ]
});
