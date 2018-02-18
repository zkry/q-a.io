<template>
  <div class="container">
    <h1>Unnamed Project</h1>
    <transition name="fadefast">
      <div class="error-msg" v-show="errorActive">
        <a class="cancel" @click="hideError">&times;</a> {{ errorMsg }}
      </div>
    </transition>
    <div class="extra-options" v-if="roomCreated" @click="closeRoom" >
      <ul>
        <li @click="closeRoom" :class="{ 'disabled-text' : roomClosed }">Close the room</li>
        <li @click="leaveRoom">Leave the room</li>
      </ul>
    </div>
    <transition name="fade" mode="out-in">
      <!-- Room Creation Page -->
      <div class="input" v-if="!roomCreated" key="roomSelect">
        <input v-model="roomName" type="text" placeholder="Enter room name" />
        <a class="go-btn" @click="onSubmit">Go</a>
      </div>

      <!-- Room Observing Page -->
      <div v-else key="roomInfo">
        <h1 style="z-index: 10;">{{ roomName }}<span class="info-text" v-if="roomClosed"> closed</span></h1>
        <h4>Questions:</h4>
        <div class='sort-btn' :class="{ 'active-text' : sortMode === 'SORTED' }" @click="toggleSort" v-if="!noQuestions">Sort</div>
        <ul class="question-container" >
          <li class="question-item" v-for="key in getQuestionKeys" :key="key" :class="{ 'closed-question' : closedRoom }">
            <span class="vote-arrow" @click="vote(questions[key].id, '1')" v-bind:class="[questions[key].myVote === '1' ? 'active-vote' : '']">▲</span>
             <span class="vote-ct" >{{ questions[key].vote }}</span>
             <span class="vote-arrow" @click="vote(questions[key].id, '-1')" v-bind:class="[questions[key].myVote === '-1' ? 'active-vote' : '']">▼</span>
             {{ questions[key].q }}
          </li>
        </ul>
        <p class="help-text" v-if="noQuestions">Have your listeners go to http://wtte.io/{{roomName}} to submit their questions and have them be voted on</p>
      </div>
    </transition>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Home',
  data: function () {
    return {
      errorMsg: '',
      errorActive: false,
      roomName: '', // ''
      roomCreated: false, // false
      userID: '',
      questions: {},
      roomClosed: false,
      sortMode: 'NORMAL'
    }
  },
  computed: {
    noQuestions () {
      console.log(this)
      return Object.keys(this.questions).length === 0
    },
    getQuestionKeys () {
      let baseKeys = Object.keys(this.questions)
      if (this.sortMode === 'NORMAL') {
        return baseKeys
      } else if (this.sortMode === 'SORTED') {
        baseKeys.sort((a, b) => {
          return this.questions[a].vote < this.questions[b].vote
        })
        return baseKeys
      } else if (this.sortMode === 'SORTED_REVERSE') {
        baseKeys.sort((a, b) => {
          return this.questions[a].vote > this.questions[b].vote
        })
        return baseKeys
      }
      return baseKeys
    }
  },
  created: function () {
    // Check if saved userID exists
    let userID = localStorage.getItem('userID')
    let roomName = localStorage.getItem('roomName')
    if (userID && roomName) {
      this.userID = localStorage.getItem('userID')
      this.roomName = localStorage.getItem('roomName')
      console.log('Reading ID from local storage:', this.userID)
      console.log('Reading room name from local storage:', roomName)
      this.roomCreated = true
      this.errorActive = false
      // Query the backend every second for new questions
      this.getQuestions()
      setInterval(() => {
        this.getQuestions()
      }, 2000)
    }
  },
  methods: {
    toggleSort () {
      if (this.sortMode === 'NORMAL') {
        this.sortMode = 'SORTED'
      } else {
        this.sortMode = 'NORMAL'
      }
    },
    leaveRoom () {
      localStorage.removeItem('userID')
      localStorage.removeItem('roomName')
      location.reload()
    },
    closeRoom () {
      const formData = {
        uID: this.userID
      }
      axios({
        method: 'post',
        url: '/api/v1/closeRoom/' + this.roomName,
        data: formData,
        params: formData,
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          this.roomClosed = true
          console.log('Room closed')
        })
        .catch(err => {
          console.log(err)
        })
    },
    hideError () {
      this.errorActive = false
    },
    showError (errorMsg) {
      this.errorMsg = errorMsg
      this.errorActive = true
    },
    getQuestions () {
      axios({
        method: 'get',
        url: '/api/v1/getQuestions/' + this.roomName,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          // TODO: Add res → this.questions
          for (let i = 0; i < res.data.questions.length; i++) {
            let id = res.data.questions[i].id
            if (!this.questions.hasOwnProperty(id)) {
              this.$set(this.questions, id, {q: res.data.questions[i].q, vote: res.data.questions[i].vote, id: id})
            }
          }
          if (res.data.is_closed) {
            this.roomClosed = true
          }
        })
    },
    onSubmit () {
      const formData = {
        roomName: this.roomName
      }
      axios({
        method: 'post',
        url: '/api/v1/createRoom',
        data: formData,
        params: formData,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          this.userID = res.data.uID
          localStorage.setItem('userID', this.userID)
          localStorage.setItem('roomName', this.roomName)
          this.roomCreated = true
          this.errorActive = false
          // Query the backend every second for new questions
          this.getQuestions()
          setInterval(() => {
            this.getQuestions()
          }, 2000)
        })
        .catch(err => {
          console.log(err.response.data)
          if (err.response.data.status === 'Bad room name') {
            this.showError('Invalid name for room. Valid characters are a thorugh z, _ and -. Room names are case insensitive')
          }
        })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
::-webkit-input-placeholder { /* Chrome/Opera/Safari */
  color: #aeaca8;
}
::-moz-placeholder { /* Firefox 19+ */
  color: #aeaca8;
}
:-ms-input-placeholder { /* IE 10+ */
  color: #aeaca8;
}
:-moz-placeholder { /* Firefox 18- */
  color: #aeaca8;
}
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  margin: 0 10px;
}
a {
  color: #42b983;
}
p {
  margin: 0;
  padding: 0;
}
input[type="text"] {
  font-size: 24px;
  text-align: center;
}

.cancel {
  color: black;
  font-weight: bold;
  cursor: pointer;
}

.error-msg {
  position: absolute;
  background-color: #e74c3c;
  top: 15px;
  right: 15px;
  padding: 5px;
  border-radius: 4px;
  cursor: grab;
  width: 200px;
}

.go-btn {
  margin-left: 10px;
  cursor: pointer;
  font-size: 24px;
}

.sort-btn::-moz-selection { background:transparent; }
.sort-btn::selection { background:transparent; }
.sort-btn {
  margin: 0 0 0 70px;
  text-align: left;
  padding: 0;
  width: 200px;
  cursor: pointer;
}

.active-text {
  color: #2980b9;
}

.container {
}
.input {
  margin: 0;
  padding: 0;
  margin-top: 100px;
}

.disabled-text {
  color: #bdc3c7;
}

.help-text {
  color: #aeaca8;
  width: 400px;
  text-align: center;
  margin: auto;
}
.question-container {
  margin: 0 50px 0 50px;
  text-align: left;
  list-style-type: none;
}

.question {
  display: block;
  white-space: nowrap;
}

.vote-arrow {
  cursor: pointer;
}

.info-text {
  font-size: x-small;
  color: grey;
}

.extra-options {
  position: absolute;
  margin-left: 0;
  padding: 10px;
  cursor: pointer;
  text-align: left;
}

.vote-ct {
  /* make width fixed ??? */
}
/* fadefast */
.fadefast-enter {
  opacity: 0;
}
.fadefast-enter-active {
  transition: opacity 1s;
}
.fadefast-leave {

}
.fadefast-leave-active {
  transition: opacity 1s;
  opacity: 0;
}

/* fade */
.fade-enter {
  opacity: 0;
}
.fade-enter-active {
  transition: opacity 1s;
  animation: slide-in 1s ease-out forwards;
}
.fade-leave {

}
.fade-leave-active {
  transition: opacity 1s;
  opacity: 0;
  animation: slide-out 1s ease-out forwards;
}

@keyframes slide-in {
  from {
    transform: translateY(20px);
  }
  to {
    transform: translateY(0px);
  }
}
@keyframes slide-out {
  from {
    transform: translateY(0px);
  }
  to {
    transform: translateY(20px);
  }
}
</style>
