import { NextResponse } from "next/server";

const PROTECTED_PATHS = ["/write", "/dashboard"];

export function middleware(request) {
  const { pathname } = request.nextUrl;
  const hasToken = request.cookies.get("mf_access_token");

  const isProtected = PROTECTED_PATHS.some((path) =>
    pathname.startsWith(path)
  );

  if (isProtected && !hasToken) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/write/:path*", "/dashboard/:path*"],
};