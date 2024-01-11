import { Link } from "react-router-dom";

export function Header() {
    return (
        <div className="header">
            <span>Dapla ctrl</span>
            <Link to="/medlemmer">Medlemmer</Link>
        </div>
    )
}