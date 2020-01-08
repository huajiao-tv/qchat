package com.huajiao.comm.common;

/**
 * Created by zhangjun-s on 16-7-25.
 */

//L:left, M:middle, R:right
public class Tuple<L,M,R> {

    private final L left;
    private final R right;
    private final M middle;

    public Tuple(L left, M middle, R right) {
        this.left = left;
        this.right = right;
        this.middle = middle;
    }

    public L getLeft() { return left; }
    public R getRight() { return right; }
    public M getMiddle() { return middle; }

    @Override
    public int hashCode() { return left.hashCode() ^ middle.hashCode() ^ right.hashCode(); }

    @Override
    public boolean equals(Object o) {
        if (!(o instanceof Tuple)) return false;
        Tuple tupleo = (Tuple) o;
        return this.left.equals(tupleo.getLeft()) &&
                this.middle.equals(tupleo.getMiddle()) &&
                this.right.equals(tupleo.getRight());
    }
}
