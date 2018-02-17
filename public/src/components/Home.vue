<template>
  <div class="container">
    <h1>Unnamed Project</h1>
    <transition name="fadefast">
      <div class="error-msg" v-show="errorActive">
        <a class="cancel" @click="hideError">&times;</a> {{ errorMsg }}
      </div>
    </transition>
    <div class="extra-options" v-if="roomCreated" @click="closeRoom" :style="closeWindowStyle">
      <span :style="closeWindowTextStyle">Close the room</span>
    </div>
    <transition name="fade" mode="out-in">
      <!-- Room Creation Page -->
      <div class="input" v-if="!roomCreated" key="roomSelect">
        <input v-model="roomName" type="text" placeholder="Enter room name" />
        <a class="go-btn" @click="onSubmit">Go</a>
      </div>

      <!-- Room Observing Page -->
      <div v-else key="roomInfo">
        <h1 style="z-index: 10;">{{ roomName }}</h1>
        <h4>Questions:</h4>
        <ul class="question-container">
          <li v-for="question in questions" :key="question.id"><span class="vote-arrow">▲</span> <span class="vote-ct">{{ question.vote }}</span> <span class="vote-arrow">▼</span> {{ question.q }}</li>
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
      questions: {},
      // Animations
      closeWindowStyle: {
        width: '200px'
      },
      closeWindowTextStyle: {
        opacity: 1
      }
    }
  },
  computed: {
    noQuestions () {
      console.log(this)
      return Object.keys(this.questions).length === 0
    }
  },
  methods: {
    closeRoom () {

    },
    closeRoomAnimation () {
      let width = 200
      let roc = 30
      let i = 0
      this.closeWindowStyle.zIndex = '-1'
      let closeAnimationID = 0
      let clearCloseAnimation = () => { clearInterval(closeAnimationID) }
      closeAnimationID = setInterval(() => {
        this.closeWindowStyle.width = width + 'px'
        this.closeWindowTextStyle.opacity -= 0.01
        i += 1
        width += roc
        if (roc <= 0) {
          clearCloseAnimation()
        }
        if (i % 20 === 0 && roc > 0) {
          roc -= 4
        }
      }, 10)
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
        url: 'http://localhost:8080/getQuestions/' + this.roomName,
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
        })
    },
    onSubmit () {
      const formData = {
        roomName: this.roomName
      }
      axios({
        method: 'post',
        url: 'http://localhost:8080/createRoom',
        data: formData,
        params: formData,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
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
.container {
}
.input {
  margin: 0;
  padding: 0;
}

.help-text {
  color: #aeaca8;
  width: 400px;
  text-align: center;
  margin: auto;
}
.question-container {
  margin: 50px;
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

.extra-options {
  position: absolute;
  margin-left: 0;
  padding: 10px;
  border-style: solid;
  border-radius: 0 4px 4px 0;
  background-color: #f1c40f;
  left: -5px;
  border-width: thin;
  cursor: pointer;
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
