import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 30,
  duration: '30s',
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
  // --- 1. LOGIN ---
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: 'test@example.com',
    password: 'password123',
  }), { headers: { 'Content-Type': 'application/json' } });

  const loginOk = check(loginRes, { 'Login Successful': (r) => r.status === 200 });

  // GUARD: If login failed, log why and STOP this iteration
  if (!loginOk) {
    console.error(`VU ${__VU} Login Failed! Status: ${loginRes.status}, Body: ${loginRes.body}`);
    return; 
  }

  sleep(1);

  // --- 2. GET USER PROFILE ---
  const meRes = http.get(`${BASE_URL}/users/me`);
  
  const meOk = check(meRes, {
    'Get Profile Successful': (r) => r.status === 200,
  });

  if (meOk) {
    const meData = meRes.json();
      console.log(meData)
    check(meRes, {
      'User ID is Present': (r) => meData.data.user!== undefined,
  
      'Workspace is Attached': (r) => meData.data.workspace !== null,
    });
  } else {
    console.error(`VU ${__VU} Profile Fetch Failed: ${meRes.status}`);
  }

  sleep(1);

  // --- 3. GET API KEYS ---
  const keysRes = http.get(`${BASE_URL}/workspaces/current/api-keys?env_id=3baac327-2345-4fdd-9f01-26ee2bf54a64`);

  check(keysRes, {

    'Get API Keys Successful': (r) => r.status === 200,
    'Is an Array': (r) => Array.isArray(r.json().data),
  });

  sleep(2);
}