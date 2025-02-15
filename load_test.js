import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
      { duration: '30s', target: 100 },
      { duration: '1m', target: 100 },
      { duration: '30s', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<50'],
      http_req_failed: ['rate<0.0001'],
    },
  };

const TOKEN = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMiwidXNlcm5hbWUiOiJ1c2VyMiIsImV4cCI6MTczOTcwMDU0Mn0.Sk_GH5t_zio6VtErzqMqn-vuPizi4x1gT7nGkpJ_Emg";

export default function () {
    let headers = { 'Authorization': TOKEN };

    // Покупка товара
    let buyRes = http.get('http://localhost:8080/api/buy/socks', { headers });
    check(buyRes, { 'buy success': (r) => r.status === 200 });

    // // Передача монет
    let payload = JSON.stringify({ toUser: "user1", amount: 1 });
    let sendRes = http.post('http://localhost:8080/api/sendCoin', payload, { headers });
    check(sendRes, { 'send success': (r) => r.status === 200 });

    sleep(0.1); 
}