import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:3000',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

export const getCount = () => api.get('/count')
export const updateCount = () => api.put('/count')

export default api
