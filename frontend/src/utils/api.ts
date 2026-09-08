const API_URL = 'http://127.0.0.1:8090/api/v1';

export async function loginUser(email, password) {
  const res = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Login failed');
  return data;
}

export async function registerUser(fullName, email, password) {
  const res = await fetch(`${API_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ full_name: fullName, email, password }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Registration failed');
  return data;
}

export async function getAlumniList(page = 1, search = '') {
  const res = await fetch(`${API_URL}/alumni?page=${page}&search=${search}`);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to fetch alumni');
  return data;
}

export async function updateProfile(token, profileData) {
  const res = await fetch(`${API_URL}/alumni/profile`, {
    method: 'PUT',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(profileData),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to update profile');
  return data;
}

export async function getJobs() {
  const res = await fetch(`${API_URL}/jobs`);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to fetch jobs');
  return data;
}

export async function getEvents() {
  const res = await fetch(`${API_URL}/events`);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to fetch events');
  return data;
}

export async function getPosts() {
  const res = await fetch(`${API_URL}/posts`);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to fetch posts');
  return data;
}

export async function createPost(formData: FormData, token: string) {
  const res = await fetch(`${API_URL}/posts`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData // Note: no Content-Type header so browser sets multipart boundary automatically
  });
  
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to create post');
  return data;
}
