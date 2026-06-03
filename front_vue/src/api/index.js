import request from './request'

export const register = (data) => request.post('/auth/register', data)
export const login = (data) => request.post('/auth/login', data)

// User Profile
export const saveProfile = (data) => request.post('/person/profile', data)
export const getProfile = (userId) => request.get('/person/profile', { params: { user_id: userId } })
export const saveWeightRecord = (data) => request.post('/person/weight-record', data)
export const getWeightRecords = (userId, limit = 30, beforeId = null) =>
  request.get('/person/weight-record', {
    params: {
      user_id: userId,
      limit,
      before_id: beforeId || undefined
    }
  })
export const deleteWeightRecord = (userId, recordId) =>
  request.delete('/person/weight-record', {
    params: {
      user_id: userId,
      record_id: recordId
    }
  })
export const getWeightDashboard = (userId, view = 'day') =>
  request.get('/person/weight-dashboard', { params: { user_id: userId, view } })
export const getAccount = (userId) => request.get('/person/account', { params: { user_id: userId } })
export const updateAccount = (data) => request.post('/person/account', data)
export const resetPassword = (data) => request.post('/person/reset-password', data)
export const getMemoryList = (userId) => request.get('/person/memory', { params: { user_id: userId } })
export const deleteMemory = (userId, memoryId) =>
  request.delete('/person/memory', {
    params: {
      user_id: userId,
      memory_id: memoryId
    }
  })

// Agent Interaction
export const askAgent = (data) => request.post('/agent/ask', data)
export const getTodayPlan = (userId) => request.get('/agent/today-plan', { params: { user_id: userId } })
export const toggleTask = (data) => request.post('/agent/toggle-task', data)

// Chat History
export const getChatMessages = (userId, chatId, pairLimit = 5, beforeId = null) =>
  request.get('/agent/chat-history', {
    params: {
      user_id: userId,
      chat_id: chatId,
      pair_limit: pairLimit,
      before_id: beforeId || undefined
    }
  })

export const deleteChatMessages = (userId, chatId) =>
  request.delete('/agent/chat-history', {
    params: {
      user_id: userId,
      chat_id: chatId
    }
  })

// Plan History
export const getPlanHistory = (userId) =>
  request.get('/agent/plan-history', {
    params: {
      user_id: userId
    }
  })

export const deletePlanHistory = (userId, recordId) =>
  request.delete('/agent/plan-history', {
    params: {
      user_id: userId,
      record_id: recordId
    }
  })

export const replacePlanFromHistory = (data) => request.post('/agent/plan-history/replace', data)

// Membership & Mall
export const getMembershipInfo = (userId) => request.get('/person/membership', { params: { user_id: userId } })
export const purchaseMembership = (data) => request.post('/person/membership/purchase', data)
export const getProducts = (category) => request.get('/person/mall/products', { params: { category } })
export const withdrawBalance = (data) => request.post('/person/balance/withdraw', data)
