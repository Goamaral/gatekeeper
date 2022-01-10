module.exports = class BaseStore {
  constructor() {
    this.store = {}
  }

  set(key, value) {
    this.store[key] = value
  }

  get(key) {
    this.store[key]
  }
}