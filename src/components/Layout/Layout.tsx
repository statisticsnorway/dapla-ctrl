import './Layout.scss'
import { Title, Button, Breadcrumb } from '@statisticsnorway/ssb-component-library'

export function Layout({title, breadcrumbItems, buttonText}: LayoutProps) {
    return (
        <div className="container">
            {breadcrumbItems?.length && <Breadcrumb items={breadcrumbItems}/>}
            <div className="title-container">
                <Title size={1}>{title}</Title>
                {buttonText && <Button>{buttonText}</Button>}
            </div>
        </div>
  )
}

interface LayoutProps {
    title: string;
    breadcrumbItems?: Array<{text: string, link?: string}>;
    buttonText?: string | ChildNode;
}