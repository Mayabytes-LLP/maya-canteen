import { createContext, useState, type FC } from "react";

export interface AppState {
  admin: boolean;
  currentPage: "canteen" | "products" | "users";
  currentUserId: number | null;
  setCurrentPage: (page: "canteen" | "products" | "users") => void;
  setCurrentUserId: (id: number) => void;
  setAdmin: (admin: boolean) => void;
}

export const initialState: AppState = {
  admin: false,
  currentPage: "canteen",
  currentUserId: null,
  setCurrentPage: () => null,
  setCurrentUserId: () => null,
  setAdmin: () => null,
};
export const AppContext = createContext<AppState>(initialState);

type Props = {
  children?: React.ReactNode;
};

export const AppProvider: FC<Props> = ({ children, ...props }) => {
  const [admin, setAdmin] = useState(initialState.admin);
  const [currentPage, setCurrentPage] = useState(initialState.currentPage);
  const [currentUserId, setCurrentUserId] = useState(
    initialState.currentUserId,
  );

  const value = {
    admin,
    currentPage,
    currentUserId,
    setCurrentPage,
    setCurrentUserId,
    setAdmin,
  };

  return (
    <AppContext.Provider {...props} value={value}>
      {children}
    </AppContext.Provider>
  );
};
