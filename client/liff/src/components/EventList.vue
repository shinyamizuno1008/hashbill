<template>
  <div>
    <h1>This is all stored events</h1>
    <h2>あなたは誰ですか</h2>
    <select @change="onChange($event)">
      <option>選んでね</option>
      <option v-for="(user, i) in users" :key="i">{{user.userName}}</option>
    </select>
    <ol class="list" v-if="selected !== ''">
      <li v-for="(event, i) in events" :key="i">
        <div class="event_detail">
          <h2>イベント名前: {{event.EventName}}</h2>
          <h2>イベント日時: {{event.Date}}</h2>
          <h2>イベント締め切り: {{event.Deadline}}</h2>
          <h2>開催場所: {{event.Location}}</h2>
          <h2>参加者上限: {{event.MembersMax}}</h2>
          <h2>抽選の有無: {{event.Lottery}}</h2>
          <h2>説明: {{event.Description}}</h2>
          <button @click="onDelete(event.HostID,event.EventName)">削除</button>
        </div>
      </li>
    </ol>
  </div>
</template>

<script>
import axios from "axios";
export default {
  data() {
    return {
      users: [],
      events: [],
      selected: ""
    };
  },
  mounted() {
    axios
      .get("/userlist", {
        headers: {
          "Access-Control-Allow-Origin": "*"
        }
      })
      .then(r => {
        this.users = r.data;
      })
      .catch(e => {
        console.log(e);
      });
  },
  methods: {
    onDelete(hostID, eventName) {
      axios
        .delete("/event/delete", {
          data: {
            hostID,
            eventName
          }
        })
        .then(r => console.log(r))
        .catch(e => console.log(e));
    },

    showUserID: function(userName) {
      if (this.users.length > 0) {
        for (let i = 0; this.users.length > i; i++) {
          if (this.users[i].userName === userName) {
            return this.users[i].userID;
          }
        }
      }
    },
    onChange(event) {
      this.selected = this.showUserID(event.target.value);
      let url = `/user/${this.selected}/eventlist`;
      axios
        .get(url)
        .then(r => {
          this.events = r.data;
        })
        .catch(e => {
          console.log(e);
        });
    }
  }
};
</script>

<style scoped>
.list {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
}
li {
  width: 600px;
}

.event_detail {
  border: 1px solid black;
}
</style>