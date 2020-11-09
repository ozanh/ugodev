<template>
  <div>
    <Playground msg="The uGO Playground" />
  </div>
</template>

<script>
import Playground from './components/Playground.vue'

export default {
  name: 'App',
  components: {
    Playground
  }
}
</script>
<style>
body {
  background-color: #3f3f3f;
  font-family: arial, helvetica, sans-serif;
}
</style>
