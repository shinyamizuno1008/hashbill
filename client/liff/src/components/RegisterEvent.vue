<template>
  <div class="container">
    <h1>イベントの登録</h1>
    <h2>あなたは誰ですか？</h2>
    <select @change="onChange($event)">
      <option>選んでね</option>
      <option v-for="(user, i) in users" :key="i">{{user.userName}}</option>
    </select>
    <form @submit.prevent="submit" action="/event/register" method="POST" ref="form">
      <div>
        <label>
          <b>主催者ID</b>
        </label>
        <input id="host" type="text" placeholder="ここにはLINEIDが入ります" :value="selected" name="hostID" />
      </div>
      <div class="form-group">
        <label for="name">
          <b>イベント名前</b>
        </label>
        <input type="text" class="form-control" id="name" name="eventName" />
      </div>
      <div class="form-group">
        <label for="date">
          <b>イベント日時</b>
        </label>
        <input type="date" class="form-control" id="eventDate" name="eventDate" />
        <input
          type="text"
          id="eventTime"
          name="eventTime"
          pattern="^([0-1]?[0-9]|2[0-4]):([0-5][0-9])$"
        />
      </div>
      <div class="form-group">
        <label for="deadline">
          <b>イベント締め切り</b>
        </label>
        <input type="date" class="form-control" id="deadlineDate" name="deadlineDate" />
        <input
          type="text"
          id="deadlineTime"
          name="deadlineTime"
          pattern="^([0-1]?[0-9]|2[0-4]):([0-5][0-9])$"
        />
      </div>
      <div class="form-group">
        <label for="location">
          <b>開催場所</b>
        </label>
        <input type="text" class="form-control" id="location" name="location" />
      </div>
      <div class="form-group">
        <label for="members-max">
          <b>参加者上限</b>
        </label>
        <input type="number" min="1" max="5" id="members-max" name="membersMax" />
      </div>
      <div class="form-group">
        <label for="lottery">
          <b>抽選の有無</b>
        </label>
        <select name="lottery">
          <option value="true">有</option>
          <option value="false">無</option>
        </select>
      </div>
      <div class="form-group">
        <textarea id="description" cols="30" rows="10" name="description"></textarea>
      </div>
      <div class="form-group">
        <button id="send" type="submit">送信</button>
      </div>
    </form>
  </div>
</template>

<script>
import Vue from "vue";
import axios from "axios";

export default {
  data: function() {
    return {
      users: [],
      selected: ""
    };
  },
  created() {
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
    submit: function() {
      this.$refs.form.submit();
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
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1,
h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
