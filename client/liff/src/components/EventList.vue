<template>
  <div>
    <h1>This is all stored events</h1>
    <ol>
      <li v-for="(event, i) in events" :key="i">{{event.EventName}}</li>
    </ol>
  </div>
</template>

<script>
import axios from "axios";
export default {
  data() {
    return {
      events: []
    };
  },
  mounted() {
    axios
      .get("/event/list")
      .then(r => {
        this.events = r.data;
      })
      .catch(e => console.log(e));
  }
};
</script>