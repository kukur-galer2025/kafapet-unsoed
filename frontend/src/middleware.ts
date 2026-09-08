import { defineMiddleware } from "astro:middleware";

export const onRequest = defineMiddleware((context, next) => {
  const { url, cookies, redirect } = context;
  
  const token = cookies.get('token')?.value;

  // Protect internal routes
  const protectedRoutes = ['/alumni', '/profile']; 
  const isProtected = protectedRoutes.some(route => url.pathname.startsWith(route));

  if (isProtected && !token) {
    return redirect('/login');
  }

  // Prevent accessing login/register if already logged in
  const guestRoutes = ['/login', '/register'];
  const isGuestRoute = guestRoutes.some(route => url.pathname.startsWith(route));

  if (isGuestRoute && token) {
    return redirect('/alumni');
  }

  return next();
});
